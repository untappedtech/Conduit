package service

import (
	"encoding/binary"
	"sync"

	"github.com/cespare/xxhash/v2"
	"github.com/untappedtech/conduit/internal/cache"
	"github.com/untappedtech/conduit/internal/domain"
)

type cachedSQL struct {
	sql  string
	args []any
}

var (
	astCacheMu     sync.RWMutex
	astCache       = cache.NewLRU[uint64, WhereExpr](1024)
	sqlCacheMu     sync.RWMutex
	sqlCache       = cache.NewLRU[uint64, cachedSQL](1024)
	cachingEnabled = true
)

// SetCachingEnabled enables or disables WHERE expression and SQL caching.
func SetCachingEnabled(enabled bool) {
	cachingEnabled = enabled
}

// ClearWhereCache purges all cached ASTs and compiled SQL queries.
func ClearWhereCache() {
	astCacheMu.Lock()
	astCache.Clear()
	astCacheMu.Unlock()

	sqlCacheMu.Lock()
	sqlCache.Clear()
	sqlCacheMu.Unlock()
}

func computeASTCacheKey(expr string, columns []domain.ColumnDef) uint64 {
	h := xxhash.New()
	_, _ = h.WriteString(expr)
	_, _ = h.WriteString("|")
	for _, col := range columns {
		_, _ = h.WriteString(col.Name)
		_, _ = h.WriteString(":")
		_, _ = h.WriteString(col.Type)
		_, _ = h.WriteString(";")
	}
	return h.Sum64()
}

func computeSQLCacheKey(expr string, columns []domain.ColumnDef, dialect SQLDialect, startParamIndex int) uint64 {
	h := xxhash.New()
	_, _ = h.WriteString(expr)
	_, _ = h.WriteString("|")
	for _, col := range columns {
		_, _ = h.WriteString(col.Name)
		_, _ = h.WriteString(":")
		_, _ = h.WriteString(col.Type)
		_, _ = h.WriteString(";")
	}
	_, _ = h.WriteString("|")
	if dialect != nil {
		_, _ = h.WriteString(dialect.QuoteIdent("x"))
		_, _ = h.WriteString(dialect.Placeholder(1))
		_, _ = h.WriteString(dialect.Placeholder(2))
	}
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], uint64(startParamIndex))
	_, _ = h.Write(buf[:])
	return h.Sum64()
}

func getCachedAST(key uint64) (WhereExpr, bool) {
	if !cachingEnabled {
		return nil, false
	}
	astCacheMu.RLock()
	defer astCacheMu.RUnlock()
	return astCache.Get(key)
}

func setCachedAST(key uint64, ast WhereExpr) {
	if !cachingEnabled || ast == nil {
		return
	}
	astCacheMu.Lock()
	defer astCacheMu.Unlock()
	astCache.Set(key, ast)
}

func getCachedSQL(key uint64) (string, []any, bool) {
	if !cachingEnabled {
		return "", nil, false
	}
	sqlCacheMu.RLock()
	defer sqlCacheMu.RUnlock()

	entry, ok := sqlCache.Get(key)
	if !ok {
		return "", nil, false
	}

	argsCopy := make([]any, len(entry.args))
	copy(argsCopy, entry.args)
	return entry.sql, argsCopy, true
}

func setCachedSQL(key uint64, sql string, args []any) {
	if !cachingEnabled {
		return
	}
	sqlCacheMu.Lock()
	defer sqlCacheMu.Unlock()

	argsCopy := make([]any, len(args))
	copy(argsCopy, args)
	sqlCache.Set(key, cachedSQL{
		sql:  sql,
		args: argsCopy,
	})
}
