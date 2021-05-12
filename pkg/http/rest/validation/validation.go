package validation

import (
	"errors"
	"path/filepath"
	"runtime"

	"github.com/KompiTech/itsm-commenting-service/pkg/validation"
)

// PayloadValidator provides payload validation
type PayloadValidator interface {
	// ValidatePayload returns error if payload is not valid
	ValidatePayload(p []byte) error
}

type payloadValidator struct {
	v validation.Validator
}

// NewPayloadValidator creates new payload validation service
func NewPayloadValidator() (PayloadValidator, error) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		return nil, errors.New("NewValidator(): could not retrieve filename from runtime.Caller()")
	}

	baseValidator := validation.NewValidator(filepath.Join(filepath.Dir(thisFile), "schema"))

	return &payloadValidator{v: baseValidator}, nil
}

// ValidatePayload returns error if payload is not valid
func (pv payloadValidator) ValidatePayload(p []byte) error {
	return pv.v.ValidateBytes(p, "add_comment.yaml")
}
