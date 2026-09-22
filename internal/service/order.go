package service

import (
	"fmt"
	"strings"

	"github.com/untappedtech/conduit/internal/domain"
)

// ValidateOrder validates that the order string refers to a valid column and has an allowed direction.
// Returns the matched canonical column name, whether sorting is descending, and an error if validation fails.
func ValidateOrder(order string, columns []domain.ColumnDef) (col string, desc bool, err error) {
	trimmed := strings.TrimSpace(order)
	if trimmed == "" {
		return "", false, nil
	}

	parts := strings.Split(trimmed, ":")
	if len(parts) > 2 {
		return "", false, &MalformedQueryError{Message: fmt.Sprintf("invalid order parameter %q, expected 'column' or 'column:asc|desc'", order)}
	}

	colInput := strings.TrimSpace(parts[0])
	if colInput == "" {
		return "", false, &MalformedQueryError{Message: "order column cannot be empty"}
	}

	var matchedCol string
	for _, c := range columns {
		if strings.EqualFold(c.Name, colInput) {
			matchedCol = c.Name
			break
		}
	}

	if matchedCol == "" {
		return "", false, &InvalidColumnError{Column: colInput}
	}

	if len(parts) == 2 {
		dir := strings.ToLower(strings.TrimSpace(parts[1]))
		switch dir {
		case "", "asc":
			desc = false
		case "desc":
			desc = true
		default:
			return "", false, &MalformedQueryError{Message: fmt.Sprintf("invalid order direction %q, must be 'asc' or 'desc'", parts[1])}
		}
	}

	return matchedCol, desc, nil
}
