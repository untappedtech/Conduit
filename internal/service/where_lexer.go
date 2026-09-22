package service

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

type tokenType int

const (
	tokEOF tokenType = iota
	tokIdent
	tokString
	tokNumber
	tokBool
	tokNull
	tokOp
	tokLParen
	tokRParen
	tokComma
)

type token struct {
	typ tokenType
	val string
	raw any
}

type lexer struct {
	input []rune
	pos   int
}

func (l *lexer) nextToken() (token, error) {
	for l.pos < len(l.input) && unicode.IsSpace(l.input[l.pos]) {
		l.pos++
	}
	if l.pos >= len(l.input) {
		return token{typ: tokEOF}, nil
	}

	ch := l.input[l.pos]

	if ch == '(' {
		l.pos++
		return token{typ: tokLParen, val: "("}, nil
	}
	if ch == ')' {
		l.pos++
		return token{typ: tokRParen, val: ")"}, nil
	}
	if ch == ',' {
		l.pos++
		return token{typ: tokComma, val: ","}, nil
	}

	// Two-character operators
	if l.pos+1 < len(l.input) {
		twoChar := string(l.input[l.pos : l.pos+2])
		if twoChar == "<=" || twoChar == ">=" || twoChar == "!=" || twoChar == "<>" || twoChar == "==" {
			l.pos += 2
			if twoChar == "==" {
				twoChar = "="
			}
			return token{typ: tokOp, val: twoChar}, nil
		}
	}

	// Single-character operators
	if ch == '=' || ch == '<' || ch == '>' {
		l.pos++
		return token{typ: tokOp, val: string(ch)}, nil
	}

	// String literal
	if ch == '\'' || ch == '"' {
		quote := ch
		l.pos++
		var sb strings.Builder
		for l.pos < len(l.input) {
			if l.input[l.pos] == quote {
				if l.pos+1 < len(l.input) && l.input[l.pos+1] == quote {
					sb.WriteRune(quote)
					l.pos += 2
					continue
				}
				l.pos++
				return token{typ: tokString, val: sb.String(), raw: sb.String()}, nil
			}
			if l.input[l.pos] == '\\' && l.pos+1 < len(l.input) {
				l.pos++
				sb.WriteRune(l.input[l.pos])
				l.pos++
				continue
			}
			sb.WriteRune(l.input[l.pos])
			l.pos++
		}
		return token{}, &MalformedQueryError{Message: "unclosed string literal in where clause"}
	}

	// Number literal
	if unicode.IsDigit(ch) || (ch == '-' && l.pos+1 < len(l.input) && unicode.IsDigit(l.input[l.pos+1])) {
		start := l.pos
		l.pos++
		isFloat := false
		for l.pos < len(l.input) && (unicode.IsDigit(l.input[l.pos]) || l.input[l.pos] == '.') {
			if l.input[l.pos] == '.' {
				isFloat = true
			}
			l.pos++
		}
		numStr := string(l.input[start:l.pos])
		if isFloat {
			val, err := strconv.ParseFloat(numStr, 64)
			if err != nil {
				return token{}, &MalformedQueryError{Message: fmt.Sprintf("invalid number %q", numStr)}
			}
			return token{typ: tokNumber, val: numStr, raw: val}, nil
		}
		val, err := strconv.ParseInt(numStr, 10, 64)
		if err != nil {
			return token{}, &MalformedQueryError{Message: fmt.Sprintf("invalid integer %q", numStr)}
		}
		return token{typ: tokNumber, val: numStr, raw: val}, nil
	}

	// Word: Identifier or Keyword
	if unicode.IsLetter(ch) || ch == '_' {
		start := l.pos
		for l.pos < len(l.input) && (unicode.IsLetter(l.input[l.pos]) || unicode.IsDigit(l.input[l.pos]) || l.input[l.pos] == '_') {
			l.pos++
		}
		word := string(l.input[start:l.pos])
		upper := strings.ToUpper(word)

		switch upper {
		case "AND", "OR", "NOT", "LIKE", "IN", "IS":
			return token{typ: tokOp, val: upper}, nil
		case "TRUE":
			return token{typ: tokBool, val: word, raw: true}, nil
		case "FALSE":
			return token{typ: tokBool, val: word, raw: false}, nil
		case "NULL":
			return token{typ: tokNull, val: word, raw: nil}, nil
		default:
			return token{typ: tokIdent, val: word}, nil
		}
	}

	return token{}, &MalformedQueryError{Message: fmt.Sprintf("unexpected character %q in where clause", ch)}
}
