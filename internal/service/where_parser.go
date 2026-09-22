package service

import (
	"fmt"
	"strings"

	"github.com/untappedtech/conduit/internal/domain"
)

type parser struct {
	tokens []token
	pos    int
	cols   []domain.ColumnDef
}

func (p *parser) peek() token {
	if p.pos < len(p.tokens) {
		return p.tokens[p.pos]
	}
	return token{typ: tokEOF}
}

func (p *parser) next() token {
	tok := p.peek()
	if p.pos < len(p.tokens) {
		p.pos++
	}
	return tok
}

func (p *parser) parseOr() (WhereExpr, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}

	for p.peek().typ == tokOp && strings.ToUpper(p.peek().val) == "OR" {
		p.next()
		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpr{Op: "OR", Left: left, Right: right}
	}

	return left, nil
}

func (p *parser) parseAnd() (WhereExpr, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}

	for p.peek().typ == tokOp && strings.ToUpper(p.peek().val) == "AND" {
		p.next()
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpr{Op: "AND", Left: left, Right: right}
	}

	return left, nil
}

func (p *parser) parseUnary() (WhereExpr, error) {
	if p.peek().typ == tokOp && strings.ToUpper(p.peek().val) == "NOT" {
		p.next()
		expr, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return &NotExpr{Expr: expr}, nil
	}

	return p.parsePrimary()
}

func (p *parser) parsePrimary() (WhereExpr, error) {
	tok := p.peek()

	if tok.typ == tokLParen {
		p.next()
		expr, err := p.parseOr()
		if err != nil {
			return nil, err
		}
		if p.peek().typ != tokRParen {
			return nil, &MalformedQueryError{Message: "expected closing parenthesis ')'"}
		}
		p.next()
		return expr, nil
	}

	if tok.typ == tokIdent {
		p.next()
		colName := tok.val

		// Validate column against schema
		var matchedCol string
		for _, c := range p.cols {
			if strings.EqualFold(c.Name, colName) {
				matchedCol = c.Name
				break
			}
		}
		if matchedCol == "" {
			return nil, &InvalidColumnError{Column: colName}
		}

		opTok := p.peek()
		if opTok.typ != tokOp {
			return nil, &MalformedQueryError{Message: fmt.Sprintf("expected operator after column %q, got %q", colName, opTok.val)}
		}
		p.next()
		op := strings.ToUpper(opTok.val)

		// Handle [NOT] IN
		if op == "NOT" {
			if p.peek().typ == tokOp && strings.ToUpper(p.peek().val) == "IN" {
				p.next()
				return p.parseInClause(matchedCol, true)
			}
			return nil, &MalformedQueryError{Message: "expected IN after NOT"}
		}
		if op == "IN" {
			return p.parseInClause(matchedCol, false)
		}

		// Handle IS [NOT] NULL
		if op == "IS" {
			isNot := false
			if p.peek().typ == tokOp && strings.ToUpper(p.peek().val) == "NOT" {
				p.next()
				isNot = true
			}
			if p.peek().typ != tokNull {
				return nil, &MalformedQueryError{Message: "expected NULL after IS"}
			}
			p.next()
			cmpOp := "IS"
			if isNot {
				cmpOp = "IS NOT"
			}
			return &ComparisonExpr{Column: matchedCol, Op: cmpOp, Value: nil}, nil
		}

		// Handle comparison operators (=, !=, <>, <, <=, >, >=, LIKE)
		switch op {
		case "=", "!=", "<>", "<", "<=", ">", ">=", "LIKE":
			valTok := p.next()
			switch valTok.typ {
			case tokString, tokNumber, tokBool:
				return &ComparisonExpr{Column: matchedCol, Op: op, Value: valTok.raw}, nil
			case tokNull:
				return &ComparisonExpr{Column: matchedCol, Op: op, Value: nil}, nil
			default:
				return nil, &MalformedQueryError{Message: fmt.Sprintf("expected literal value after operator %q, got %q", op, valTok.val)}
			}
		default:
			return nil, &MalformedQueryError{Message: fmt.Sprintf("unsupported operator %q", op)}
		}
	}

	return nil, &MalformedQueryError{Message: fmt.Sprintf("unexpected token %q in where clause", tok.val)}
}

func (p *parser) parseInClause(column string, not bool) (WhereExpr, error) {
	if p.peek().typ != tokLParen {
		return nil, &MalformedQueryError{Message: "expected '(' after IN"}
	}
	p.next()

	var values []any
	for {
		valTok := p.next()
		switch valTok.typ {
		case tokString, tokNumber, tokBool:
			values = append(values, valTok.raw)
		case tokNull:
			values = append(values, nil)
		default:
			return nil, &MalformedQueryError{Message: fmt.Sprintf("expected literal value in IN list, got %q", valTok.val)}
		}

		if p.peek().typ == tokComma {
			p.next()
			continue
		}
		if p.peek().typ == tokRParen {
			p.next()
			break
		}
		return nil, &MalformedQueryError{Message: "expected ',' or ')' in IN list"}
	}

	return &InExpr{Column: column, Not: not, Values: values}, nil
}

// ParseWhere parses a WHERE expression string into an AST, validating column names against table schema.
func ParseWhere(expr string, columns []domain.ColumnDef) (WhereExpr, error) {
	trimmed := strings.TrimSpace(expr)
	if trimmed == "" {
		return nil, nil
	}

	l := &lexer{input: []rune(trimmed)}
	var tokens []token
	for {
		tok, err := l.nextToken()
		if err != nil {
			return nil, err
		}
		if tok.typ == tokEOF {
			break
		}
		tokens = append(tokens, tok)
	}

	if len(tokens) == 0 {
		return nil, nil
	}

	p := &parser{tokens: tokens, cols: columns}
	ast, err := p.parseOr()
	if err != nil {
		return nil, err
	}
	if p.pos < len(p.tokens) {
		return nil, &MalformedQueryError{Message: fmt.Sprintf("unexpected extra token %q in where clause", p.tokens[p.pos].val)}
	}

	return ast, nil
}
