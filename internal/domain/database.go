package domain

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"reflect"
)

var (
	ErrPrimaryKeyMissing = errors.New("table does not expose a single primary key column and is non-addressable by ID")
)

type ColumnDef struct {
	Name          string `json:"name" yaml:"name" xml:"name" toml:"name"`
	Type          string `json:"type" yaml:"type" xml:"type" toml:"type"`
	Nullable      *bool  `json:"nullable,omitempty" yaml:"nullable,omitempty" xml:"nullable,omitempty" toml:"nullable,omitempty"`
	Unique        *bool  `json:"unique,omitempty" yaml:"unique,omitempty" xml:"unique,omitempty" toml:"unique,omitempty"`
	Default       any    `json:"default,omitempty" yaml:"default,omitempty" xml:"default,omitempty" toml:"default,omitempty"`
	PK            *bool  `json:"pk,omitempty" yaml:"pk,omitempty" xml:"pk,omitempty" toml:"pk,omitempty"`
	Autoincrement *bool  `json:"autoincrement,omitempty" yaml:"autoincrement,omitempty" xml:"autoincrement,omitempty" toml:"autoincrement,omitempty"`
	CID           *int   `json:"cid,omitempty" yaml:"cid,omitempty" xml:"cid,omitempty" toml:"cid,omitempty"`
}

func (c ColumnDef) DefaultValue() any {
	if c.Default == nil {
		return nil
	}
	v := reflect.ValueOf(c.Default)
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil
		}
		return v.Elem().Interface()
	}
	return c.Default
}

func (c ColumnDef) DefaultString() string {
	val := c.DefaultValue()
	if val == nil {
		return ""
	}
	return fmt.Sprint(val)
}

type XMLSchema struct {
	XMLName xml.Name    `xml:"schema"`
	Columns []ColumnDef `xml:"column"`
}

type XMLTableList struct {
	XMLName xml.Name `xml:"tables"`
	Tables  []string `xml:"table"`
}

type ListRequest struct {
	Limit  int
	Offset int
	Order  string
	Where  string
}

type DatabaseDriver interface {
	Schema(ctx context.Context, tableName string) ([]ColumnDef, error)
	ListTables(ctx context.Context) ([]string, error)
	CreateTable(ctx context.Context, tableName string, columns []ColumnDef) error
	DropTable(ctx context.Context, tableName string) error

	List(ctx context.Context, tableName string, req ListRequest) ([]map[string]any, error)
	GetByID(ctx context.Context, tableName string, recordID string) (map[string]any, error)
	Insert(ctx context.Context, tableName string, recordData map[string]any) (map[string]any, error)
	Update(ctx context.Context, tableName string, recordID string, recordData map[string]any) (map[string]any, error)
	Delete(ctx context.Context, tableName string, recordID string) error
	HealthCheck(ctx context.Context) error
	Close() error
}
