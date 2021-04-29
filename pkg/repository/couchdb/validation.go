package couchdb

import (
	"encoding/json"
	"path"
	"runtime"

	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment"
	"github.com/KompiTech/rmap"
	"github.com/pkg/errors"
)

// Validator provides comment validation
type Validator interface {
	Validate(c comment.Comment) error
}

type validator struct {
}

// NewValidator creates new validation service
func NewValidator() Validator {
	return &validator{}
}

// Validate returns error if comment is not valid
func (v validator) Validate(c comment.Comment) error {
	js, err := json.Marshal(c)
	if err != nil {
		return err
	}

	asset, err := rmap.NewFromBytes(js)
	if err != nil {
		return err
	}

	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		return errors.Wrap(err, "could not retrieve filename from runtime.Caller()")
	}

	filepath := path.Join(path.Dir(filename), "./schema/comment.yaml")
	sf, err := rmap.NewFromYAMLFile(filepath)
	if err != nil {
		return err
	}

	err = asset.ValidateSchema(sf)
	if err != nil {
		return err
	}

	return nil
}
