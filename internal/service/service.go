package service

import (
	"context"

	"github.com/untappedtech/conduit/internal/domain"
)

type APIService struct {
	databaseDriver domain.DatabaseDriver
	serverConfig   *domain.ServerConfig
}

func NewAPIService(databaseDriver domain.DatabaseDriver, serverConfig *domain.ServerConfig) *APIService {
	return &APIService{databaseDriver: databaseDriver, serverConfig: serverConfig}
}

func (apiService *APIService) ListTables(ctx context.Context) ([]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return apiService.databaseDriver.ListTables(ctx)
}

func (apiService *APIService) GetSchema(ctx context.Context, tableName string) ([]domain.ColumnDef, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return apiService.databaseDriver.Schema(ctx, tableName)
}

func (apiService *APIService) CreateTable(ctx context.Context, tableName string, columns []domain.ColumnDef) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return apiService.databaseDriver.CreateTable(ctx, tableName, columns)
}

func (apiService *APIService) DropTable(ctx context.Context, tableName string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return apiService.databaseDriver.DropTable(ctx, tableName)
}

func (apiService *APIService) List(ctx context.Context, tableName string, req domain.ListRequest) ([]map[string]any, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if req.Order != "" || req.Where != "" {
		schema, err := apiService.databaseDriver.Schema(ctx, tableName)
		if err != nil {
			return nil, err
		}
		if req.Order != "" {
			if _, _, err := ValidateOrder(req.Order, schema); err != nil {
				return nil, err
			}
		}
		if req.Where != "" {
			if _, err := ParseWhere(req.Where, schema); err != nil {
				return nil, err
			}
		}
	}

	defaultLimit := apiService.serverConfig.Server.DefaultLimit

	// Case 1: limit missing or explicitly 0 → use default limit
	if req.Limit <= 0 {
		req.Limit = defaultLimit
	} else {
		// Case 2: limit > 0 → apply server cap if present
		if defaultLimit > 0 && req.Limit > defaultLimit {
			req.Limit = defaultLimit
		}
	}

	// Case 3: effective limit = 0 → unlimited
	if req.Limit == 0 {
		req.Limit = -1 // sentinel for unlimited
	}

	return apiService.databaseDriver.List(ctx, tableName, req)
}

func (apiService *APIService) GetByID(ctx context.Context, tableName string, recordID string) (map[string]any, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return apiService.databaseDriver.GetByID(ctx, tableName, recordID)
}

func (apiService *APIService) Insert(ctx context.Context, tableName string, recordData map[string]any) (map[string]any, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return apiService.databaseDriver.Insert(ctx, tableName, recordData)
}

func (apiService *APIService) Update(ctx context.Context, tableName string, recordID string, recordData map[string]any) (map[string]any, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return apiService.databaseDriver.Update(ctx, tableName, recordID, recordData)
}

func (apiService *APIService) Delete(ctx context.Context, tableName string, recordID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return apiService.databaseDriver.Delete(ctx, tableName, recordID)
}
