package repository

import (
	uuidgen "github.com/google/uuid"
	"github.com/pkg/errors"
)

// GenerateUUID returns a random UUID
func GenerateUUID() (string, error) {
	uuid, err := uuidgen.NewRandom()
	if err != nil {
		return "", errors.Wrap(err, "Could not generate UUID")
	}

	return uuid.String(), nil
}
