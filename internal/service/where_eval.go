package service

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

func (b *BinaryExpr) Eval(row map[string]any) bool {
	if b.Op == "AND" {
		return b.Left.Eval(row) && b.Right.Eval(row)
	}
	return b.Left.Eval(row) || b.Right.Eval(row)
}

func (n *NotExpr) Eval(row map[string]any) bool {
	return !n.Expr.Eval(row)
}

func (c *ComparisonExpr) Eval(row map[string]any) bool {
	rowVal := findRowVal(row, c.Column)

	if c.Value == nil {
		if c.Op == "!=" || c.Op == "<>" || c.Op == "IS NOT" {
			return rowVal != nil
		}
		return rowVal == nil
	}
	if rowVal == nil {
		return false
	}

	if strings.ToUpper(c.Op) == "LIKE" {
		pattern, ok := c.Value.(string)
		if !ok {
			return false
		}
		rowStr := fmt.Sprintf("%v", rowVal)
		return matchLike(pattern, rowStr)
	}

	cmp, ok := compareValues(rowVal, c.Value)
	if !ok {
		return false
	}

	switch c.Op {
	case "=", "==":
		return cmp == 0
	case "!=", "<>":
		return cmp != 0
	case "<":
		return cmp < 0
	case "<=":
		return cmp <= 0
	case ">":
		return cmp > 0
	case ">=":
		return cmp >= 0
	default:
		return false
	}
}

func (in *InExpr) Eval(row map[string]any) bool {
	rowVal := findRowVal(row, in.Column)
	if rowVal == nil {
		return in.Not
	}

	found := false
	for _, v := range in.Values {
		if cmp, ok := compareValues(rowVal, v); ok && cmp == 0 {
			found = true
			break
		}
	}

	if in.Not {
		return !found
	}
	return found
}

func findRowVal(row map[string]any, col string) any {
	if val, ok := row[col]; ok {
		return val
	}
	for k, val := range row {
		if strings.EqualFold(k, col) {
			return val
		}
	}
	return nil
}

func matchLike(pattern, s string) bool {
	// Convert SQL LIKE pattern (% -> .*, _ -> .) to regex
	var sb strings.Builder
	sb.WriteString("(?i)^")
	runes := []rune(pattern)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		switch r {
		case '%':
			sb.WriteString(".*")
		case '_':
			sb.WriteString(".")
		default:
			sb.WriteString(regexp.QuoteMeta(string(r)))
		}
	}
	sb.WriteString("$")
	re, err := regexp.Compile(sb.String())
	if err != nil {
		return false
	}
	return re.MatchString(s)
}

// CompareValues compares two values with dynamic type coercion.
// Returns -1 for a < b, 0 for a == b, 1 for a > b, and ok = true if comparable.
func CompareValues(a, b any) (int, bool) {
	// Number comparison
	aNum, aIsNum := toFloat64(a)
	bNum, bIsNum := toFloat64(b)
	if aIsNum && bIsNum {
		if aNum < bNum {
			return -1, true
		} else if aNum > bNum {
			return 1, true
		}
		return 0, true
	}

	// Boolean comparison
	aBool, aIsBool := toBool(a)
	bBool, bIsBool := toBool(b)
	if aIsBool && bIsBool {
		if aBool == bBool {
			return 0, true
		}
		if !aBool && bBool {
			return -1, true
		}
		return 1, true
	}

	// String comparison
	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)
	return strings.Compare(aStr, bStr), true
}

func compareValues(a, b any) (int, bool) {
	return CompareValues(a, b)
}

func toFloat64(v any) (float64, bool) {
	switch n := v.(type) {
	case int:
		return float64(n), true
	case int8:
		return float64(n), true
	case int16:
		return float64(n), true
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	case uint:
		return float64(n), true
	case uint8:
		return float64(n), true
	case uint16:
		return float64(n), true
	case uint32:
		return float64(n), true
	case uint64:
		return float64(n), true
	case float32:
		return float64(n), true
	case float64:
		return n, true
	case string:
		if f, err := strconv.ParseFloat(n, 64); err == nil {
			return f, true
		}
	}
	return 0, false
}

func toBool(v any) (bool, bool) {
	switch b := v.(type) {
	case bool:
		return b, true
	case string:
		if lower := strings.ToLower(b); lower == "true" {
			return true, true
		} else if lower == "false" {
			return false, true
		}
	}
	return false, false
}
