package impl

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	_ "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/untappedtech/conduit/internal/domain"
	"github.com/untappedtech/conduit/internal/service"
)

var validClickHouseIdent = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

type ClickHouseEngine struct {
	sqlDatabase *sql.DB
}

func NewClickHouseEngine(dataSourceName string) (domain.DatabaseDriver, error) {
	dbConn, err := sql.Open("clickhouse", dataSourceName)
	if err != nil {
		return nil, err
	}
	if err := dbConn.Ping(); err != nil {
		return nil, err
	}
	return &ClickHouseEngine{sqlDatabase: dbConn}, nil
}

func (engine *ClickHouseEngine) QuoteIdent(identifier string) string {
	return "`" + strings.ReplaceAll(identifier, "`", "``") + "`"
}

func (engine *ClickHouseEngine) Placeholder(index int) string {
	return "?"
}

func (engine *ClickHouseEngine) Schema(ctx context.Context, tableName string) ([]domain.ColumnDef, error) {
	if !validClickHouseIdent.MatchString(tableName) {
		return nil, domain.ErrInvalidID
	}

	query := `SELECT name, type, is_in_primary_key, default_expression 
		FROM system.columns 
		WHERE database = currentDatabase() AND table = ? 
		ORDER BY position;`

	rows, err := engine.sqlDatabase.QueryContext(ctx, query, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columnDefinitions []domain.ColumnDef
	cid := 0
	for rows.Next() {
		var colName, colType string
		var isPKNum uint8
		var defaultExpr sql.NullString

		if err := rows.Scan(&colName, &colType, &isPKNum, &defaultExpr); err != nil {
			return nil, err
		}

		lowerType := strings.ToLower(colType)
		nullable := strings.HasPrefix(lowerType, "nullable(")
		isPK := isPKNum == 1
		cidValue := cid

		col := domain.ColumnDef{
			Name:     colName,
			Type:     colType,
			Nullable: &nullable,
			PK:       &isPK,
			CID:      &cidValue,
		}
		if defaultExpr.Valid && defaultExpr.String != "" {
			val := defaultExpr.String
			col.Default = &val
		}

		columnDefinitions = append(columnDefinitions, col)
		cid++
	}

	if len(columnDefinitions) == 0 {
		return nil, domain.ErrNotFound
	}
	return columnDefinitions, nil
}

func (engine *ClickHouseEngine) detectPK(ctx context.Context, tableName string) (string, error) {
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

func (engine *ClickHouseEngine) ListTables(ctx context.Context) ([]string, error) {
	query := `SELECT name FROM system.tables WHERE database = currentDatabase() AND is_temporary = 0 ORDER BY name;`
	rows, err := engine.sqlDatabase.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err == nil {
			tables = append(tables, name)
		}
	}
	return tables, nil
}

func mapClickHouseType(t string) string {
	switch strings.ToLower(strings.TrimSpace(t)) {
	case "integer", "int":
		return "Int64"
	case "text", "string", "varchar":
		return "String"
	case "real", "float", "double":
		return "Float64"
	case "boolean", "bool":
		return "Bool"
	default:
		return t
	}
}

func (engine *ClickHouseEngine) CreateTable(ctx context.Context, tableName string, columns []domain.ColumnDef) error {
	if !validClickHouseIdent.MatchString(tableName) {
		return domain.ErrInvalidID
	}
	if len(columns) == 0 {
		return fmt.Errorf("at least one column required")
	}

	var columnDeclarations []string
	var pkCols []string

	for _, col := range columns {
		if !validClickHouseIdent.MatchString(col.Name) {
			return fmt.Errorf("invalid column name: %s", col.Name)
		}
		chType := mapClickHouseType(col.Type)
		if col.Nullable != nil && *col.Nullable && !strings.HasPrefix(strings.ToLower(chType), "nullable(") {
			chType = fmt.Sprintf("Nullable(%s)", chType)
		}

		colSQL := fmt.Sprintf("%s %s", engine.QuoteIdent(col.Name), chType)
		if col.PK != nil && *col.PK {
			pkCols = append(pkCols, engine.QuoteIdent(col.Name))
		}
		columnDeclarations = append(columnDeclarations, colSQL)
	}

	orderByClause := "ORDER BY tuple()"
	if len(pkCols) > 0 {
		orderByClause = fmt.Sprintf("ORDER BY (%s)", strings.Join(pkCols, ", "))
	}

	query := fmt.Sprintf(
		"CREATE TABLE IF NOT EXISTS %s (%s) ENGINE = MergeTree() %s;",
		engine.QuoteIdent(tableName),
		strings.Join(columnDeclarations, ", "),
		orderByClause,
	)

	_, err := engine.sqlDatabase.ExecContext(ctx, query)
	return err
}

func (engine *ClickHouseEngine) DropTable(ctx context.Context, tableName string) error {
	if !validClickHouseIdent.MatchString(tableName) {
		return domain.ErrInvalidID
	}
	query := fmt.Sprintf("DROP TABLE IF EXISTS %s;", engine.QuoteIdent(tableName))
	_, err := engine.sqlDatabase.ExecContext(ctx, query)
	return err
}

func (engine *ClickHouseEngine) List(ctx context.Context, tableName string, req domain.ListRequest) ([]map[string]any, error) {
	if !validClickHouseIdent.MatchString(tableName) {
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

func (engine *ClickHouseEngine) GetByID(ctx context.Context, tableName string, recordID string) (map[string]any, error) {
	pkCol, err := engine.detectPK(ctx, tableName)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf("SELECT * FROM %s WHERE %s = ? LIMIT 1", engine.QuoteIdent(tableName), engine.QuoteIdent(pkCol))
	rows, err := engine.sqlDatabase.QueryContext(ctx, query, recordID)
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

func (engine *ClickHouseEngine) Insert(ctx context.Context, tableName string, recordData map[string]any) (map[string]any, error) {
	columnNames := make([]string, 0, len(recordData))
	placeholderMarks := make([]string, 0, len(recordData))
	valuesList := make([]any, 0, len(recordData))

	for key, val := range recordData {
		columnNames = append(columnNames, engine.QuoteIdent(key))
		placeholderMarks = append(placeholderMarks, "?")
		valuesList = append(valuesList, val)
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s);", engine.QuoteIdent(tableName), strings.Join(columnNames, ", "), strings.Join(placeholderMarks, ", "))
	_, err := engine.sqlDatabase.ExecContext(ctx, query, valuesList...)
	if err != nil {
		return nil, err
	}

	return recordData, nil
}

func (engine *ClickHouseEngine) Update(ctx context.Context, tableName string, recordID string, recordData map[string]any) (map[string]any, error) {
	pkCol, err := engine.detectPK(ctx, tableName)
	if err != nil {
		return nil, err
	}

	setAssignments := make([]string, 0, len(recordData))
	valuesList := make([]any, 0, len(recordData)+1)

	for key, val := range recordData {
		setAssignments = append(setAssignments, fmt.Sprintf("%s = ?", engine.QuoteIdent(key)))
		valuesList = append(valuesList, val)
	}
	valuesList = append(valuesList, recordID)

	query := fmt.Sprintf("ALTER TABLE %s UPDATE %s WHERE %s = ?;", engine.QuoteIdent(tableName), strings.Join(setAssignments, ", "), engine.QuoteIdent(pkCol))
	_, err = engine.sqlDatabase.ExecContext(ctx, query, valuesList...)
	if err != nil {
		return nil, err
	}

	return engine.GetByID(ctx, tableName, recordID)
}

func (engine *ClickHouseEngine) Delete(ctx context.Context, tableName string, recordID string) error {
	pkCol, err := engine.detectPK(ctx, tableName)
	if err != nil {
		return err
	}

	query := fmt.Sprintf("ALTER TABLE %s DELETE WHERE %s = ?;", engine.QuoteIdent(tableName), engine.QuoteIdent(pkCol))
	_, err = engine.sqlDatabase.ExecContext(ctx, query, recordID)
	return err
}

func (engine *ClickHouseEngine) HealthCheck(ctx context.Context) error {
	return engine.sqlDatabase.PingContext(ctx)
}

func (engine *ClickHouseEngine) Close() error {
	return engine.sqlDatabase.Close()
}

