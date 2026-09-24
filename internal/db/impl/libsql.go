package impl

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
	"github.com/untappedtech/conduit/internal/domain"
	"github.com/untappedtech/conduit/internal/service"
)

var validLibSQLIdent = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

type LibSQLEngine struct {
	sqlDatabase *sql.DB
}

func NewLibSQLEngine(dataSourceName string) (domain.DatabaseDriver, error) {
	dbConn, err := sql.Open("libsql", dataSourceName)
	if err != nil {
		return nil, err
	}
	if err := dbConn.Ping(); err != nil {
		return nil, err
	}
	return &LibSQLEngine{sqlDatabase: dbConn}, nil
}

func (engine *LibSQLEngine) QuoteIdent(identifier string) string {
	return `"` + strings.ReplaceAll(identifier, `"`, `""`) + `"`
}

func (engine *LibSQLEngine) Placeholder(index int) string {
	return "?"
}

func (engine *LibSQLEngine) Schema(ctx context.Context, tableName string) ([]domain.ColumnDef, error) {
	if !validLibSQLIdent.MatchString(tableName) {
		return nil, domain.ErrInvalidID
	}

	schemaQuery := fmt.Sprintf(`PRAGMA table_info(%s)`, engine.QuoteIdent(tableName))
	rows, err := engine.sqlDatabase.QueryContext(ctx, schemaQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ddlSQL sql.NullString
	_ = engine.sqlDatabase.QueryRowContext(ctx, "SELECT sql FROM sqlite_master WHERE type='table' AND name=?", tableName).Scan(&ddlSQL)
	hasAutoIncrement := ddlSQL.Valid && strings.Contains(strings.ToUpper(ddlSQL.String), "AUTOINCREMENT")

	var columnDefinitions []domain.ColumnDef
	for rows.Next() {
		var columnID int
		var columnName string
		var columnType string
		var notNullFlag int
		var defaultVal sql.NullString
		var primaryKeyFlag int

		if err := rows.Scan(&columnID, &columnName, &columnType, &notNullFlag, &defaultVal, &primaryKeyFlag); err != nil {
			return nil, err
		}

		nullableFlag := notNullFlag == 0
		isPK := primaryKeyFlag > 0
		col := domain.ColumnDef{
			Name:     columnName,
			Type:     columnType,
			Nullable: &nullableFlag,
			PK:       &isPK,
			CID:      &columnID,
		}
		if defaultVal.Valid {
			strVal := defaultVal.String
			col.Default = &strVal
		}
		if isPK && hasAutoIncrement {
			isAuto := true
			col.Autoincrement = &isAuto
		}

		columnDefinitions = append(columnDefinitions, col)
	}

	if len(columnDefinitions) == 0 {
		return nil, domain.ErrNotFound
	}

	return columnDefinitions, nil
}

func (engine *LibSQLEngine) detectPK(ctx context.Context, tableName string) (string, error) {
	columns, err := engine.Schema(ctx, tableName)
	if err != nil {
		return "", err
	}

	pkCount := 0
	pkName := ""
	for _, col := range columns {
		if col.PK != nil && *col.PK {
			pkCount++
			pkName = col.Name
		}
	}
	if pkCount != 1 {
		return "", domain.ErrPrimaryKeyMissing
	}
	return pkName, nil
}

func (engine *LibSQLEngine) ListTables(ctx context.Context) ([]string, error) {
	tablesQuery := `SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' AND name NOT LIKE 'auth_%' ORDER BY name;`
	rows, err := engine.sqlDatabase.QueryContext(ctx, tablesQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tableNames []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err == nil {
			tableNames = append(tableNames, tableName)
		}
	}
	return tableNames, nil
}

func (engine *LibSQLEngine) CreateTable(ctx context.Context, tableName string, columns []domain.ColumnDef) error {
	if !validLibSQLIdent.MatchString(tableName) {
		return domain.ErrInvalidID
	}
	if len(columns) == 0 {
		return fmt.Errorf("at least one column required")
	}

	dbTransaction, err := engine.sqlDatabase.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer dbTransaction.Rollback()

	var columnDeclarations []string
	for _, col := range columns {
		if !validLibSQLIdent.MatchString(col.Name) {
			return fmt.Errorf("invalid column name: %s", col.Name)
		}
		colSQL := fmt.Sprintf("%s %s", engine.QuoteIdent(col.Name), strings.ToUpper(col.Type))

		isPK := col.PK != nil && *col.PK
		isAuto := col.Autoincrement != nil && *col.Autoincrement

		if isAuto {
			colSQL = fmt.Sprintf("%s INTEGER PRIMARY KEY AUTOINCREMENT", engine.QuoteIdent(col.Name))
		} else if isPK {
			colSQL += " PRIMARY KEY"
		}

		if col.Nullable != nil && !*col.Nullable && !isPK {
			colSQL += " NOT NULL"
		}
		if col.Unique != nil && *col.Unique && !isPK {
			colSQL += " UNIQUE"
		}
		if col.Default != nil {
			colSQL += fmt.Sprintf(" DEFAULT %s", *col.Default)
		}

		columnDeclarations = append(columnDeclarations, colSQL)
	}

	createQuery := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (%s);`, engine.QuoteIdent(tableName), strings.Join(columnDeclarations, ", "))
	if _, err := dbTransaction.ExecContext(ctx, createQuery); err != nil {
		return err
	}

	return dbTransaction.Commit()
}

func (engine *LibSQLEngine) DropTable(ctx context.Context, tableName string) error {
	if !validLibSQLIdent.MatchString(tableName) {
		return domain.ErrInvalidID
	}
	dropQuery := fmt.Sprintf(`DROP TABLE IF EXISTS %s;`, engine.QuoteIdent(tableName))
	_, err := engine.sqlDatabase.ExecContext(ctx, dropQuery)
	return err
}

func (engine *LibSQLEngine) List(ctx context.Context, tableName string, req domain.ListRequest) ([]map[string]any, error) {
	if !validLibSQLIdent.MatchString(tableName) {
		return nil, domain.ErrInvalidID
	}

	query := fmt.Sprintf("SELECT * FROM %s", engine.QuoteIdent(tableName))
	args := []any{}

	if req.Where != "" {
		cols, err := engine.Schema(ctx, tableName)
		if err != nil {
			return nil, err
		}
		whereSQL, whereArgs, err := service.ParseWhereSQL(req.Where, cols, engine, 1)
		if err != nil {
			return nil, err
		}
		if whereSQL != "" {
			query += " WHERE " + whereSQL
			args = append(args, whereArgs...)
		}
	}

	if req.Order != "" {
		cols, err := engine.Schema(ctx, tableName)
		if err != nil {
			return nil, err
		}
		col, desc, err := service.ValidateOrder(req.Order, cols)
		if err != nil {
			return nil, err
		}
		if col != "" {
			if desc {
				query += fmt.Sprintf(" ORDER BY %s DESC", engine.QuoteIdent(col))
			} else {
				query += fmt.Sprintf(" ORDER BY %s ASC", engine.QuoteIdent(col))
			}
		}
	}

	// Unlimited mode: req.Limit < 0 → no LIMIT clause
	if req.Limit < 0 {
		query += " OFFSET ?"
		args = append(args, req.Offset)
	} else {
		// Normal bounded mode
		query += " LIMIT ? OFFSET ?"
		args = append(args, req.Limit, req.Offset)
	}

	rows, err := engine.sqlDatabase.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanRowsToMapSlice(rows)
}

func (engine *LibSQLEngine) GetByID(ctx context.Context, tableName string, recordID string) (map[string]any, error) {
	pkCol, err := engine.detectPK(ctx, tableName)
	if err != nil {
		return nil, err
	}

	selectQuery := fmt.Sprintf(`SELECT * FROM %s WHERE %s = ? LIMIT 1`, engine.QuoteIdent(tableName), engine.QuoteIdent(pkCol))
	rows, err := engine.sqlDatabase.QueryContext(ctx, selectQuery, recordID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results, err := scanRowsToMapSlice(rows)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, domain.ErrNotFound
	}
	return results[0], nil
}

func (engine *LibSQLEngine) Insert(ctx context.Context, tableName string, recordData map[string]any) (map[string]any, error) {
	columnNames := make([]string, 0, len(recordData))
	placeholderMarks := make([]string, 0, len(recordData))
	valuesList := make([]any, 0, len(recordData))

	for key, val := range recordData {
		columnNames = append(columnNames, engine.QuoteIdent(key))
		placeholderMarks = append(placeholderMarks, "?")
		valuesList = append(valuesList, val)
	}

	insertQuery := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s)`, engine.QuoteIdent(tableName), strings.Join(columnNames, ", "), strings.Join(placeholderMarks, ", "))
	result, err := engine.sqlDatabase.ExecContext(ctx, insertQuery, valuesList...)
	if err != nil {
		return nil, err
	}

	pkCol, err := engine.detectPK(ctx, tableName)
	if err != nil {
		return recordData, nil
	}

	lastInsertID, err := result.LastInsertId()
	if err == nil && lastInsertID > 0 {
		return engine.GetByID(ctx, tableName, fmt.Sprintf("%d", lastInsertID))
	}

	if pkVal, ok := recordData[pkCol]; ok {
		return engine.GetByID(ctx, tableName, fmt.Sprintf("%v", pkVal))
	}

	return recordData, nil
}

func (engine *LibSQLEngine) Update(ctx context.Context, tableName string, recordID string, recordData map[string]any) (map[string]any, error) {
	pkCol, err := engine.detectPK(ctx, tableName)
	if err != nil {
		return nil, err
	}

	setAssignments := make([]string, 0, len(recordData))
	valuesList := make([]any, 0, len(recordData)+1)

	for key, val := range recordData {
		setAssignments = append(setAssignments, fmt.Sprintf(`%s = ?`, engine.QuoteIdent(key)))
		valuesList = append(valuesList, val)
	}
	valuesList = append(valuesList, recordID)

	updateQuery := fmt.Sprintf(`UPDATE %s SET %s WHERE %s = ?`, engine.QuoteIdent(tableName), strings.Join(setAssignments, ", "), engine.QuoteIdent(pkCol))
	_, err = engine.sqlDatabase.ExecContext(ctx, updateQuery, valuesList...)
	if err != nil {
		return nil, err
	}

	return engine.GetByID(ctx, tableName, recordID)
}

func (engine *LibSQLEngine) Delete(ctx context.Context, tableName string, recordID string) error {
	pkCol, err := engine.detectPK(ctx, tableName)
	if err != nil {
		return err
	}

	deleteQuery := fmt.Sprintf(`DELETE FROM %s WHERE %s = ?`, engine.QuoteIdent(tableName), engine.QuoteIdent(pkCol))
	_, err = engine.sqlDatabase.ExecContext(ctx, deleteQuery, recordID)
	return err
}

func (engine *LibSQLEngine) HealthCheck(ctx context.Context) error {
	return engine.sqlDatabase.PingContext(ctx)
}

func (engine *LibSQLEngine) Close() error {
	return engine.sqlDatabase.Close()
}

