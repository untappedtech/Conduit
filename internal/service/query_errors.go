package service

import "fmt"

// InvalidColumnError represents an error when a query references a column not present in the table schema.
type InvalidColumnError struct {
	Column string
}

func (e *InvalidColumnError) Error() string {
	return fmt.Sprintf("invalid column name: %q", e.Column)
}

// IsInvalidColumn returns true if the error is an InvalidColumnError.
func IsInvalidColumn(err error) bool {
	_, ok := err.(*InvalidColumnError)
	return ok
}

// MalformedQueryError represents a syntax or semantic error in a query expression.
type MalformedQueryError struct {
	Message string
}

func (e *MalformedQueryError) Error() string {
	return fmt.Sprintf("malformed query syntax: %s", e.Message)
}

// IsMalformedQuery returns true if the error is a MalformedQueryError.
func IsMalformedQuery(err error) bool {
	_, ok := err.(*MalformedQueryError)
	return ok
}
