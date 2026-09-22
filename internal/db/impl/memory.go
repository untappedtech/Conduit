package impl

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/untappedtech/conduit/internal/domain"
	"github.com/untappedtech/conduit/internal/service"
)

type MemoryDB struct {
	mutex     sync.RWMutex
	tablesMap map[string]*MemoryTable
}

type MemoryTable struct {
	Columns []domain.ColumnDef
	Rows    []map[string]any
}

func NewMemoryDB() domain.DatabaseDriver {
	return &MemoryDB{
		tablesMap: make(map[string]*MemoryTable),
	}
}

func (memoryDatabase *MemoryDB) CreateTable(ctx context.Context, tableName string, columns []domain.ColumnDef) error {
	memoryDatabase.mutex.Lock()
	defer memoryDatabase.mutex.Unlock()

	if _, tableExists := memoryDatabase.tablesMap[tableName]; tableExists {
		return errors.New("table already exists")
	}

	orderedColumns := make([]domain.ColumnDef, len(columns))
	for index, col := range columns {
		cidValue := index
		orderedColumns[index] = domain.ColumnDef{
			Name:          col.Name,
			Type:          col.Type,
			Nullable:      col.Nullable,
			Unique:        col.Unique,
			Default:       col.Default,
			PK:            col.PK,
			Autoincrement: col.Autoincrement,
			CID:           &cidValue,
		}
	}

	memoryDatabase.tablesMap[tableName] = &MemoryTable{
		Columns: orderedColumns,
		Rows:    []map[string]any{},
	}

	return nil
}

func (memoryDatabase *MemoryDB) detectPK(tableName string) (string, error) {
	targetTable, tableExists := memoryDatabase.tablesMap[tableName]
	if !tableExists {
		return "", domain.ErrNotFound
	}
	pkCount := 0
	pkName := ""
	for _, col := range targetTable.Columns {
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

func (memoryDatabase *MemoryDB) GetByID(ctx context.Context, tableName string, recordID string) (map[string]any, error) {
	memoryDatabase.mutex.RLock()
	defer memoryDatabase.mutex.RUnlock()

	pkCol, err := memoryDatabase.detectPK(tableName)
	if err != nil {
		return nil, err
	}

	targetTable := memoryDatabase.tablesMap[tableName]
	for _, currentRow := range targetTable.Rows {
		if idValue, ok := currentRow[pkCol]; ok && fmt.Sprintf("%v", idValue) == recordID {
			clonedRow := make(map[string]any)
			for key, val := range currentRow {
				clonedRow[key] = val
			}
			return clonedRow, nil
		}
	}

	return nil, domain.ErrNotFound
}

func (memoryDatabase *MemoryDB) List(ctx context.Context, tableName string, req domain.ListRequest) ([]map[string]any, error) {
	memoryDatabase.mutex.RLock()
	defer memoryDatabase.mutex.RUnlock()

	targetTable, tableExists := memoryDatabase.tablesMap[tableName]
	if !tableExists {
		return nil, domain.ErrNotFound
	}

	var matchedRows []map[string]any
	if req.Where != "" {
		expr, err := service.ParseWhere(req.Where, targetTable.Columns)
		if err != nil {
			return nil, err
		}
		if expr != nil {
			for _, row := range targetTable.Rows {
				if expr.Eval(row) {
					matchedRows = append(matchedRows, row)
				}
			}
		} else {
			matchedRows = targetTable.Rows
		}
	} else {
		matchedRows = targetTable.Rows
	}

	outputRows := make([]map[string]any, len(matchedRows))
	for index, currentRow := range matchedRows {
		clonedRow := make(map[string]any, len(currentRow))
		for key, val := range currentRow {
			clonedRow[key] = val
		}
		outputRows[index] = clonedRow
	}

	if req.Order != "" {
		col, desc, err := service.ValidateOrder(req.Order, targetTable.Columns)
		if err != nil {
			return nil, err
		}
		if col != "" {
			sort.SliceStable(outputRows, func(i, j int) bool {
				valI := outputRows[i][col]
				valJ := outputRows[j][col]
				if valI == nil && valJ == nil {
					return false
				}
				if valI == nil {
					return !desc
				}
				if valJ == nil {
					return desc
				}
				cmp, ok := service.CompareValues(valI, valJ)
				if !ok {
					return false
				}
				if desc {
					return cmp > 0
				}
				return cmp < 0
			})
		}
	}

	// Offset beyond range → empty slice
	if req.Offset >= len(outputRows) {
		return []map[string]any{}, nil
	}

	// Unlimited mode: limit < 0 means "return all rows starting at offset"
	if req.Limit < 0 {
		return outputRows[req.Offset:], nil
	}

	// Normal bounded mode
	endIndex := req.Offset + req.Limit
	if endIndex > len(outputRows) {
		endIndex = len(outputRows)
	}

	return outputRows[req.Offset:endIndex], nil
}

func (memoryDatabase *MemoryDB) Update(ctx context.Context, tableName string, recordID string, recordData map[string]any) (map[string]any, error) {
	memoryDatabase.mutex.Lock()
	defer memoryDatabase.mutex.Unlock()

	pkCol, err := memoryDatabase.detectPK(tableName)
	if err != nil {
		return nil, err
	}

	targetTable := memoryDatabase.tablesMap[tableName]
	for index, currentRow := range targetTable.Rows {
		if idValue, ok := currentRow[pkCol]; ok && fmt.Sprintf("%v", idValue) == recordID {
			for key, val := range recordData {
				currentRow[key] = val
			}
			targetTable.Rows[index] = currentRow
			clonedRow := make(map[string]any)
			for key, val := range currentRow {
				clonedRow[key] = val
			}
			return clonedRow, nil
		}
	}

	return nil, domain.ErrNotFound
}

func (memoryDatabase *MemoryDB) Delete(ctx context.Context, tableName string, recordID string) error {
	memoryDatabase.mutex.Lock()
	defer memoryDatabase.mutex.Unlock()

	pkCol, err := memoryDatabase.detectPK(tableName)
	if err != nil {
		return err
	}

	targetTable := memoryDatabase.tablesMap[tableName]
	for index, currentRow := range targetTable.Rows {
		if idValue, ok := currentRow[pkCol]; ok && fmt.Sprintf("%v", idValue) == recordID {
			targetTable.Rows = append(targetTable.Rows[:index], targetTable.Rows[index+1:]...)
			return nil
		}
	}

	return domain.ErrNotFound
}

func (memoryDatabase *MemoryDB) Insert(ctx context.Context, tableName string, recordData map[string]any) (map[string]any, error) {
	memoryDatabase.mutex.Lock()
	defer memoryDatabase.mutex.Unlock()

	targetTable, tableExists := memoryDatabase.tablesMap[tableName]
	if !tableExists {
		return nil, domain.ErrNotFound
	}

	if pkCol, pkErr := memoryDatabase.detectPK(tableName); pkErr == nil {
		if _, exists := recordData[pkCol]; !exists {
			var newAutoID int64 = int64(len(targetTable.Rows) + 1)
			recordData[pkCol] = newAutoID
		}
	}

	targetTable.Rows = append(targetTable.Rows, recordData)

	clonedRow := make(map[string]any)
	for key, val := range recordData {
		clonedRow[key] = val
	}
	return clonedRow, nil
}

func (memoryDatabase *MemoryDB) Schema(ctx context.Context, tableName string) ([]domain.ColumnDef, error) {
	memoryDatabase.mutex.RLock()
	defer memoryDatabase.mutex.RUnlock()

	targetTable, tableExists := memoryDatabase.tablesMap[tableName]
	if !tableExists {
		return nil, domain.ErrNotFound
	}

	clonedColumns := make([]domain.ColumnDef, len(targetTable.Columns))
	copy(clonedColumns, targetTable.Columns)
	return clonedColumns, nil
}

func (memoryDatabase *MemoryDB) ListTables(ctx context.Context) ([]string, error) {
	memoryDatabase.mutex.RLock()
	defer memoryDatabase.mutex.RUnlock()

	tableNames := make([]string, 0, len(memoryDatabase.tablesMap))
	for name := range memoryDatabase.tablesMap {
		tableNames = append(tableNames, name)
	}

	return tableNames, nil
}

func (memoryDatabase *MemoryDB) DropTable(ctx context.Context, tableName string) error {
	memoryDatabase.mutex.Lock()
	defer memoryDatabase.mutex.Unlock()

	if _, tableExists := memoryDatabase.tablesMap[tableName]; !tableExists {
		return domain.ErrNotFound
	}

	delete(memoryDatabase.tablesMap, tableName)
	return nil
}

func (memoryDatabase *MemoryDB) HealthCheck(ctx context.Context) error {
	return nil
}

func (memoryDatabase *MemoryDB) Close() error {
	return nil
}
