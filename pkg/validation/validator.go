package validation

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"github.com/qri-io/jsonschema"
	"gopkg.in/yaml.v2"
)

// ErrGeneral represents general error returned during pre-validation setup
type ErrGeneral struct {
	err error
}

// NewErrGeneral returns wrapped error
func NewErrGeneral(err error) error {
	return &ErrGeneral{
		err: errors.Wrap(err, "general error"),
	}
}

// Error returns error message
func (e *ErrGeneral) Error() string {
	return e.err.Error()
}

// Validator provides comment validation
type Validator interface {
	ValidateBytes(b []byte, schemaFile string) error
}

type validator struct {
	schemaDir   string
	schemaFiles embed.FS
}

// NewValidator creates new validation service
func NewValidator(schemaFiles embed.FS) Validator {
	return &validator{
		schemaDir:   "schema",
		schemaFiles: schemaFiles,
	}
}

// ValidateBytes returns error if b is not valid
func (v validator) ValidateBytes(b []byte, schemaFile string) error {
	fp := filepath.Join(v.schemaDir, schemaFile)

	file, err := v.schemaFiles.Open(fp)
	if err != nil {
		return NewErrGeneral(errors.Wrapf(err, "could not open file %s", fp))
	}

	var schemaData interface{}
	if err := yaml.NewDecoder(file).Decode(&schemaData); err != nil {
		return NewErrGeneral(errors.Wrapf(err, "could not decode schema definition file %s", schemaFile))
	}

	schemaData = convert(schemaData)

	var schemaBytes []byte
	if schemaBytes, err = json.Marshal(schemaData); err != nil {
		return NewErrGeneral(errors.Wrap(err, "could not marshal schema data"))
	}

	rs := &jsonschema.Schema{}
	if err := json.Unmarshal(schemaBytes, rs); err != nil {
		return NewErrGeneral(errors.Wrap(err, "could not unmarshal schema bytes"))
	}

	errs, err := rs.ValidateBytes(context.Background(), b)
	if err != nil {
		return err // parsing error
	}

	if len(errs) > 0 { // schema validation errors
		// join errors into one
		errorStrings := make([]string, 0)
		for _, err := range errs {
			errorStrings = append(errorStrings,
				strings.ReplaceAll(fmt.Sprintf("%s: %s", err.PropertyPath, err.Message), `"`, "'"))
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
