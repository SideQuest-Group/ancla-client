package ancla

import (
	"errors"
	"fmt"
)

// APIError represents an error response from the Ancla API.
type APIError struct {
	StatusCode int
	Message    string
}

// Error implements the error interface.
func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("ancla api: %d %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("ancla api: %d", e.StatusCode)
}

// IsNotFound reports whether the error is a 404 Not Found response.
func IsNotFound(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == 404
	}
	return false
}

// IsUnauthorized reports whether the error is a 401 Unauthorized response.
func IsUnauthorized(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == 401
	}
	return false
}

// IsForbidden reports whether the error is a 403 Forbidden response.
func IsForbidden(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == 403
	}
	return false
}
