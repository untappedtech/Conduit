package db

import (
	"errors"
	"strings"

	"github.com/untappedtech/conduit/internal/db/impl"
	"github.com/untappedtech/conduit/internal/domain"
)

func NewDatabase(serverConfig *domain.ServerConfig) (domain.DatabaseDriver, error) {
	var rawEngine domain.DatabaseDriver
	var err error

	driver := strings.ToLower(strings.TrimSpace(serverConfig.Database.Driver))

	switch driver {
	case "sqlite", "sqlite3":
		rawEngine, err = impl.NewSQLiteEngine(serverConfig.Database.DSN)
	case "memory", "mem", "in-memory", "inmemory":
		rawEngine = impl.NewMemoryDB()
	case "postgres", "postgresql", "pgsql", "cockroach", "cockroachdb":
		rawEngine, err = impl.NewPostgresEngine(serverConfig.Database.DSN)
	case "mysql", "mariadb", "tidb":
		rawEngine, err = impl.NewMySQLEngine(serverConfig.Database.DSN)
	case "sqlserver", "mssql", "microsoftsqlserver":
		rawEngine, err = impl.NewSQLServerEngine(serverConfig.Database.DSN)
	case "libsql", "turso":
		rawEngine, err = impl.NewLibSQLEngine(serverConfig.Database.DSN)
	case "clickhouse", "ch":
		rawEngine, err = impl.NewClickHouseEngine(serverConfig.Database.DSN)
	case "oracle", "ora":
		rawEngine, err = impl.NewOracleEngine(serverConfig.Database.DSN)
	default:
		return nil, errors.New("unsupported database driver: " + serverConfig.Database.Driver)
	}

	if err != nil {
		return nil, err
	}

	return NewCachedDatabase(rawEngine, 256), nil
}
