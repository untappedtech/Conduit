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

// Error implements the error interface.
func (e ErrorSpec) Error() string {
	if len(e.Details) > 0 {
		return fmt.Sprintf("%s: %s (%s)", e.Title, e.DefaultMessage, strings.Join(e.Details, "; "))
	}
	return fmt.Sprintf("%s: %s", e.Title, e.DefaultMessage)
}

// Clone creates an independent deep copy of the ErrorSpec.
func (e ErrorSpec) Clone() ErrorSpec {
	var details []string
	if len(e.Details) > 0 {
		details = make([]string, len(e.Details))
		copy(details, e.Details)
	}
	return ErrorSpec{
		Title:          e.Title,
		Status:         e.Status,
		DefaultMessage: e.DefaultMessage,
		Details:        details,
	}
}

// With returns a new ErrorSpec instance with the provided error and contextual details attached.
// The receiver is never modified, guaranteeing safe concurrent use without data leakage.
func (e ErrorSpec) With(err error, context ...string) ErrorSpec {
	instance := e.Clone()
	instance.Details = append(instance.Details, BuildDetails(err, context...)...)
	return instance
}

// WithDetails returns a new ErrorSpec instance with contextual details attached.
func (e ErrorSpec) WithDetails(context ...string) ErrorSpec {
	return e.With(nil, context...)
}
