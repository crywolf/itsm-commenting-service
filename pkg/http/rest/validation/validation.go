package validation

import (
	"embed"

	"github.com/KompiTech/itsm-commenting-service/pkg/validation"
)

//go:embed schema
var schemaFiles embed.FS

// PayloadValidator provides payload validation
type PayloadValidator interface {
	// ValidatePayload returns error if payload is not valid
	ValidatePayload(p []byte, schemaFile string) error
}

type payloadValidator struct {
	v validation.Validator
}

// NewPayloadValidator creates new payload validation service
func NewPayloadValidator() (PayloadValidator, error) {
	baseValidator := validation.NewValidator(schemaFiles)

	return &payloadValidator{v: baseValidator}, nil
}

// ValidatePayload returns error if payload is not valid
func (pv payloadValidator) ValidatePayload(p []byte, schemaFile string) error {
	return pv.v.ValidateBytes(p, schemaFile)
}
