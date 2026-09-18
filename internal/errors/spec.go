package errors

import (
	"fmt"
	"strings"
)

// ErrorSpec defines a structured RFC-style error type.
type ErrorSpec struct {
	Title          string
	Status         int
	DefaultMessage string
	Details        []string
}

// Slug converts an error title into a documentation slug.
func (e ErrorSpec) Slug() string {
	slug := strings.ToLower(e.Title)
	return strings.ReplaceAll(slug, " ", "-")
}

// DocumentationURL builds the full documentation URL for an error.
func (e ErrorSpec) DocumentationURL() string {
	return fmt.Sprintf("https://conduit.untapped.tech/docs/errors/%s", e.Slug())
}

// Attach adds contextual details or error messages to the spec.
func (e *ErrorSpec) Attach(err error, context ...string) {
	if err != nil {
		e.Details = append(e.Details, fmt.Sprintf("Error: %v", err))
	}
	for _, c := range context {
		if c != "" {
			e.Details = append(e.Details, c)
		}
	}
}

// Reset clears details so the spec can be reused safely.
func (e *ErrorSpec) Reset() {
	e.Details = nil
}
