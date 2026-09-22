package errors

import (
	"net/http"
)

// Centralized catalog
var (
	ErrBadRequest = ErrorSpec{
		Title:          "Bad Request",
		Status:         http.StatusBadRequest,
		DefaultMessage: "The request could not be processed.",
	}

	ErrUnauthorized = ErrorSpec{
		Title:          "Unauthorized",
		Status:         http.StatusUnauthorized,
		DefaultMessage: "Authentication is required.",
	}

	ErrForbidden = ErrorSpec{
		Title:          "Forbidden",
		Status:         http.StatusForbidden,
		DefaultMessage: "You do not have permission to perform this action.",
	}

	ErrNotFound = ErrorSpec{
		Title:          "Not Found",
		Status:         http.StatusNotFound,
		DefaultMessage: "The requested resource does not exist.",
	}

	ErrMethodNotAllowed = ErrorSpec{
		Title:          "Method Not Allowed",
		Status:         http.StatusMethodNotAllowed,
		DefaultMessage: "This HTTP method is not allowed for this endpoint.",
	}

	ErrConflict = ErrorSpec{
		Title:          "Conflict",
		Status:         http.StatusConflict,
		DefaultMessage: "A conflict occurred while processing the request.",
	}

	ErrUnsupportedMediaType = ErrorSpec{
		Title:          "Unsupported Media Type",
		Status:         http.StatusUnsupportedMediaType,
		DefaultMessage: "The request format is not supported.",
	}

	ErrUnprocessableEntity = ErrorSpec{
		Title:          "Unprocessable Entity",
		Status:         http.StatusUnprocessableEntity,
		DefaultMessage: "The request could not be validated.",
	}

	ErrTooManyRequests = ErrorSpec{
		Title:          "Too Many Requests",
		Status:         http.StatusTooManyRequests,
		DefaultMessage: "Rate limit exceeded.",
	}

	ErrInternalServerError = ErrorSpec{
		Title:          "Internal Server Error",
		Status:         http.StatusInternalServerError,
		DefaultMessage: "An unexpected error occurred.",
	}

	ErrNotImplemented = ErrorSpec{
		Title:          "Not Implemented",
		Status:         http.StatusNotImplemented,
		DefaultMessage: "This feature is not implemented.",
	}

	ErrServiceUnavailable = ErrorSpec{
		Title:          "Service Unavailable",
		Status:         http.StatusServiceUnavailable,
		DefaultMessage: "The service is temporarily unavailable.",
	}
)
