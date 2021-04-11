package couchdb

import (
	"net/http"

	"github.com/KompiTech/itsm-commenting-service/pkg/repository"
)

// ErrorNorFound returns an error with the supplied message and HTTP 404 status code
func ErrorNorFound(message string) error {
	return repository.NewError(message, http.StatusNotFound)
}

// ErrorConflict returns an error with the supplied message and HTTP 409 status code
func ErrorConflict(message string) error {
	return repository.NewError(message, http.StatusConflict)
}
