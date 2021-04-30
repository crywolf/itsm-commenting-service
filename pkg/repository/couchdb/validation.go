package couchdb

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"runtime"
	"sort"
	"strings"

	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment"
	"github.com/pkg/errors"
	"github.com/qri-io/jsonschema"
	"gopkg.in/yaml.v2"
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
	commentJSON, err := json.Marshal(c)
	if err != nil {
		return err
	}

	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		return errors.New("could not retrieve filename from runtime.Caller()")
	}

	filepath := path.Join(path.Dir(filename), "./schema/comment.yaml")

	file, err := os.Open(filepath)
	if err != nil {
		return errors.Wrapf(err, "could not open file %s", filepath)
	}

	var schemaData interface{}
	if err := yaml.NewDecoder(file).Decode(&schemaData); err != nil {
		return errors.Wrap(err, "could not decode schema definition file")
	}

	schemaData = convert(schemaData)

	var schemaBytes []byte
	if schemaBytes, err = json.Marshal(schemaData); err != nil {
		return errors.Wrap(err, "could not marshal schema data")
	}

	rs := &jsonschema.Schema{}
	if err := json.Unmarshal(schemaBytes, rs); err != nil {
		return errors.Wrap(err, "could not unmarshal schema bytes")
	}

	errs, err := rs.ValidateBytes(context.Background(), commentJSON)
	if err != nil {
		return errors.Wrap(err, "could not validate asset bytes")
	}

	if len(errs) > 0 {
		// join errors into one
		errorStrings := make([]string, len(errs))
		for _, err := range errs {
			errorStrings = append(errorStrings, fmt.Sprintf("%s: %s", err.PropertyPath, err.Message))
		}
		sort.Strings(errorStrings)

		return errors.New(strings.Join(errorStrings, "\n"))
	}

	return nil
}

// convert recursively converts each encountered map[interface{}]interface{} to a map[string]interface{} value
func convert(i interface{}) interface{} {
	switch x := i.(type) {
	case map[interface{}]interface{}:
		m2 := map[string]interface{}{}
		for k, v := range x {
			m2[k.(string)] = convert(v)
		}
		return m2
	case []interface{}:
		for i, v := range x {
			x[i] = convert(v)
		}
	}
	return i
}
