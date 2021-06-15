package couchdb

import (
	"embed"
	"encoding/json"

	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment"
	"github.com/KompiTech/itsm-commenting-service/pkg/validation"
)

//go:embed schema
var schemaFiles embed.FS

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
	baseValidator := validation.NewValidator(schemaFiles)

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
