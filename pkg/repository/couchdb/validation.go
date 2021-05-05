package couchdb

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"runtime"

	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment"
	"github.com/KompiTech/itsm-commenting-service/pkg/validation"
)

// Validator provides comment validation
type Validator interface {
	// Validate returns error if comment is not valid
	Validate(c comment.Comment) error
}

type validator struct {
	v validation.Validator
}

// NewValidator creates new validation service
func NewValidator() (Validator, error) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		return nil, errors.New("NewValidator(): could not retrieve filename from runtime.Caller()")
	}

	baseValidator := validation.NewValidator(filepath.Join(filepath.Dir(thisFile), "schema"))

	return &validator{v: baseValidator}, nil
}

// Validate returns error if comment is not valid
func (r validator) Validate(c comment.Comment) error {
	commentJSON, err := json.Marshal(c)
	if err != nil {
		return err
	}

	return r.v.ValidateBytes(commentJSON, "comment.yaml")
}
