package service

import (
	"fmt"
	"strings"

	"github.com/untappedtech/conduit/internal/domain"
)

// SQLDialect defines how a SQL database quotes identifiers and renders parameter placeholders.
// Database implementations provide their own dialect without requiring modifications to the service package.
type SQLDialect interface {
	QuoteIdent(name string) string
	Placeholder(index int) string
}

// SimpleDialect allows constructing an ad-hoc or test SQLDialect using functions.
type SimpleDialect struct {
	QuoteIdentFunc  func(name string) string
	PlaceholderFunc func(index int) string
}

func (s SimpleDialect) QuoteIdent(name string) string {
	if s.QuoteIdentFunc != nil {
		return s.QuoteIdentFunc(name)
	}
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

func (s SimpleDialect) Placeholder(index int) string {
	if s.PlaceholderFunc != nil {
		return s.PlaceholderFunc(index)
	}
	return "?"
}

type ansiDialect struct{}

func (ansiDialect) QuoteIdent(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

func (ansiDialect) Placeholder(index int) string {
	return "?"
}

func ensureDialect(dialect SQLDialect) SQLDialect {
	if dialect == nil {
		return ansiDialect{}
	}
	return dialect
}

func (b *BinaryExpr) ToSQL(dialect SQLDialect, startParamIndex int) (string, []any) {
	dialect = ensureDialect(dialect)
	leftSQL, leftArgs := b.Left.ToSQL(dialect, startParamIndex)
	rightSQL, rightArgs := b.Right.ToSQL(dialect, startParamIndex+len(leftArgs))
	sql := fmt.Sprintf("(%s %s %s)", leftSQL, b.Op, rightSQL)
	return sql, append(leftArgs, rightArgs...)
}

func (n *NotExpr) ToSQL(dialect SQLDialect, startParamIndex int) (string, []any) {
	dialect = ensureDialect(dialect)
	sql, args := n.Expr.ToSQL(dialect, startParamIndex)
	return fmt.Sprintf("NOT (%s)", sql), args
}

func (c *ComparisonExpr) ToSQL(dialect SQLDialect, startParamIndex int) (string, []any) {
	dialect = ensureDialect(dialect)
	quotedCol := dialect.QuoteIdent(c.Column)
	if c.Value == nil {
		if c.Op == "!=" || c.Op == "<>" || c.Op == "IS NOT" {
			return fmt.Sprintf("%s IS NOT NULL", quotedCol), nil
		}
		return fmt.Sprintf("%s IS NULL", quotedCol), nil
	}

	op := c.Op
	if op == "<>" {
		op = "!="
	}

	placeholder := dialect.Placeholder(startParamIndex)
	return fmt.Sprintf("%s %s %s", quotedCol, op, placeholder), []any{c.Value}
}

func (in *InExpr) ToSQL(dialect SQLDialect, startParamIndex int) (string, []any) {
	dialect = ensureDialect(dialect)
	quotedCol := dialect.QuoteIdent(in.Column)
	if len(in.Values) == 0 {
		if in.Not {
			return "1=1", nil
		}
		return "1=0", nil
	}

	placeholders := make([]string, len(in.Values))
	for i := range in.Values {
		placeholders[i] = dialect.Placeholder(startParamIndex + i)
	}

	op := "IN"
	if in.Not {
		op = "NOT IN"
	}

	sql := fmt.Sprintf("%s %s (%s)", quotedCol, op, strings.Join(placeholders, ", "))
	return sql, in.Values
}

// ParseWhereSQL parses a WHERE clause and produces safe SQL and parameter arguments for the given dialect.
func ParseWhereSQL(expr string, columns []domain.ColumnDef, dialect SQLDialect, startParamIndex int) (string, []any, error) {
	ast, err := ParseWhere(expr, columns)
	if err != nil {
		return "", nil, err
	}
	if ast == nil {
		return "", nil, nil
	}
	sql, args := ast.ToSQL(dialect, startParamIndex)
	return sql, args, nil
}
