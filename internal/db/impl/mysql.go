package impl

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/untappedtech/conduit/internal/domain"
	"github.com/untappedtech/conduit/internal/service"
)

var validMySQLIdent = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

type MySQLEngine struct {
	sqlDatabase *sql.DB
}

func NewMySQLEngine(dataSourceName string) (domain.DatabaseDriver, error) {
	dbConn, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		return nil, err
	}
	if err := dbConn.Ping(); err != nil {
		return nil, err
	}
	return &MySQLEngine{sqlDatabase: dbConn}, nil
}

func (engine *MySQLEngine) QuoteIdent(identifier string) string {
	return "`" + strings.ReplaceAll(identifier, "`", "``") + "`"
}

func (engine *MySQLEngine) Placeholder(index int) string {
	return "?"
}

func (engine *MySQLEngine) Schema(ctx context.Context, tableName string) ([]domain.ColumnDef, error) {
	if !validMySQLIdent.MatchString(tableName) {
		return nil, domain.ErrInvalidID
	}

	query := fmt.Sprintf(`DESCRIBE %s;`, engine.QuoteIdent(tableName))
	rows, err := engine.sqlDatabase.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columnDefinitions []domain.ColumnDef
	cid := 0
	for rows.Next() {
		var fieldName, colType, nullAllowed, keyType sql.NullString
		var defaultValue, extraInfo sql.NullString

		if err := rows.Scan(&fieldName, &colType, &nullAllowed, &keyType, &defaultValue, &extraInfo); err != nil {
			return nil, err
		}

		nullableFlag := nullAllowed.String == "YES"
		isPK := strings.ToUpper(keyType.String) == "PRI"
		cidValue := cid
		col := domain.ColumnDef{
			Name:     fieldName.String,
			Type:     colType.String,
			Nullable: &nullableFlag,
			PK:       &isPK,
			CID:      &cidValue,
		}
		if defaultValue.Valid {
			strVal := defaultValue.String
			col.Default = &strVal
		}

		isAuto := strings.Contains(strings.ToLower(extraInfo.String), "auto_increment")
		if isAuto {
			col.Autoincrement = &isAuto
		}

		columnDefinitions = append(columnDefinitions, col)
		cid++
	}

	if len(columnDefinitions) == 0 {
		return nil, domain.ErrNotFound
	}
	return columnDefinitions, nil
}

func (engine *MySQLEngine) detectPK(ctx context.Context, tableName string) (string, error) {
	cols, err := engine.Schema(ctx, tableName)
	if err != nil {
		return "", err
	}
	pkCount := 0
	pkName := ""
	for _, col := range cols {
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

func (engine *MySQLEngine) ListTables(ctx context.Context) ([]string, error) {
	rows, err := engine.sqlDatabase.QueryContext(ctx, `SHOW TABLES;`)
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

func (engine *MySQLEngine) CreateTable(ctx context.Context, tableName string, columns []domain.ColumnDef) error {
	if !validMySQLIdent.MatchString(tableName) {
		return domain.ErrInvalidID
	}
	if len(columns) == 0 {
		return fmt.Errorf("at least one column required")
	}

	var columnDeclarations []string
	for _, col := range columns {
		if !validMySQLIdent.MatchString(col.Name) {
			return fmt.Errorf("invalid column name: %s", col.Name)
		}
		colSQL := fmt.Sprintf("%s %s", engine.QuoteIdent(col.Name), strings.ToUpper(col.Type))

		if col.PK != nil && *col.PK {
			if col.Autoincrement != nil && *col.Autoincrement {
				colSQL += " PRIMARY KEY AUTO_INCREMENT"
			} else {
				colSQL += " PRIMARY KEY"
			}
		}
		if col.Nullable != nil && !*col.Nullable {
			colSQL += " NOT NULL"
		}
		columnDeclarations = append(columnDeclarations, colSQL)
	}

	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (%s);`, engine.QuoteIdent(tableName), strings.Join(columnDeclarations, ", "))
	_, err := engine.sqlDatabase.ExecContext(ctx, query)
	return err
}

func (engine *MySQLEngine) DropTable(ctx context.Context, tableName string) error {
	if !validMySQLIdent.MatchString(tableName) {
		return domain.ErrInvalidID
	}
	_, err := engine.sqlDatabase.ExecContext(ctx, fmt.Sprintf(`DROP TABLE IF EXISTS %s;`, engine.QuoteIdent(tableName)))
	return err
}

func (engine *MySQLEngine) List(ctx context.Context, tableName string, req domain.ListRequest) ([]map[string]any, error) {
	if !validMySQLIdent.MatchString(tableName) {
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

func (engine *MySQLEngine) GetByID(ctx context.Context, tableName string, recordID string) (map[string]any, error) {
	pkCol, err := engine.detectPK(ctx, tableName)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf(`SELECT * FROM %s WHERE %s = ? LIMIT 1`, engine.QuoteIdent(tableName), engine.QuoteIdent(pkCol))
	rows, err := engine.sqlDatabase.QueryContext(ctx, query, recordID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results, err := scanRowsToMapSlice(rows)
	if err != nil || len(results) == 0 {
		return nil, domain.ErrNotFound
	}
	return results[0], nil
}

func (engine *MySQLEngine) Insert(ctx context.Context, tableName string, recordData map[string]any) (map[string]any, error) {
	columnNames := make([]string, 0, len(recordData))
	placeholderMarks := make([]string, 0, len(recordData))
	valuesList := make([]any, 0, len(recordData))

	for key, val := range recordData {
		columnNames = append(columnNames, engine.QuoteIdent(key))
		placeholderMarks = append(placeholderMarks, "?")
		valuesList = append(valuesList, val)
	}

	query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s);`, engine.QuoteIdent(tableName), strings.Join(columnNames, ", "), strings.Join(placeholderMarks, ", "))
	result, err := engine.sqlDatabase.ExecContext(ctx, query, valuesList...)
	if err != nil {
		return nil, err
	}

	lastInsertedID, err := result.LastInsertId()
	if err == nil && lastInsertedID > 0 {
		if _, pkErr := engine.detectPK(ctx, tableName); pkErr == nil {
			return engine.GetByID(ctx, tableName, fmt.Sprintf("%d", lastInsertedID))
		}
	}
	return recordData, nil
}

func (engine *MySQLEngine) Update(ctx context.Context, tableName string, recordID string, recordData map[string]any) (map[string]any, error) {
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

	query := fmt.Sprintf(`UPDATE %s SET %s WHERE %s = ?;`, engine.QuoteIdent(tableName), strings.Join(setAssignments, ", "), engine.QuoteIdent(pkCol))
	_, err = engine.sqlDatabase.ExecContext(ctx, query, valuesList...)
	if err != nil {
		return nil, err
	}

	return engine.GetByID(ctx, tableName, recordID)
}

func (engine *MySQLEngine) Delete(ctx context.Context, tableName string, recordID string) error {
	pkCol, err := engine.detectPK(ctx, tableName)
	if err != nil {
		return err
	}

	query := fmt.Sprintf(`DELETE FROM %s WHERE %s = ?;`, engine.QuoteIdent(tableName), engine.QuoteIdent(pkCol))
	_, err = engine.sqlDatabase.ExecContext(ctx, query, recordID)
	return err
}

func (engine *MySQLEngine) HealthCheck(ctx context.Context) error {
	return engine.sqlDatabase.PingContext(ctx)
}

func (engine *MySQLEngine) Close() error {
	return engine.sqlDatabase.Close()
}
