package service

// WhereExpr represents an AST node for a WHERE filter expression.
type WhereExpr interface {
	ToSQL(dialect SQLDialect, startParamIndex int) (string, []any)
	Eval(row map[string]any) bool
}

// BinaryExpr represents an AND or OR logical combination.
type BinaryExpr struct {
	Op    string // "AND" or "OR"
	Left  WhereExpr
	Right WhereExpr
}

// NotExpr represents a logical negation.
type NotExpr struct {
	Expr WhereExpr
}

// ComparisonExpr represents column <op> value comparison.
type ComparisonExpr struct {
	Column string
	Op     string // "=", "!=", "<>", "<", "<=", ">", ">=", "LIKE", "IS", "IS NOT"
	Value  any
}

// InExpr represents column [NOT] IN (val1, val2, ...).
type InExpr struct {
	Column string
	Not    bool
	Values []any
}
