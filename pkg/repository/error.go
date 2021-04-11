package repository

import "net/http"

// Error represents an error returned by repository
type Error struct {
	message    string
	httpStatus int
}

// NewError returns an error with the supplied message and HTTP status code
func NewError(message string, httpStatus int) error {
	return &Error{
		message:    message,
		httpStatus: httpStatus,
	}
}

// StatusCode returns the HTTP status code associated with the error,
// or 500 (internal server error), if none
func (e *Error) StatusCode() int {
	if e.httpStatus == 0 {
		return http.StatusInternalServerError
	}
	return e.httpStatus
}

// Error returns error message
func (e *Error) Error() string {
	return e.message
}
