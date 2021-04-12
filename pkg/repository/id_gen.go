package repository

import (
	"io"

	uuidgen "github.com/google/uuid"
	"github.com/pkg/errors"
)

// GenerateUUID returns a random UUID
func GenerateUUID(rand io.Reader) (string, error) {
	uuidgen.SetRand(rand)

	uuid, err := uuidgen.NewRandom()
	if err != nil {
		return "", errors.Wrap(err, "Could not generate UUID")
	}

	return uuid.String(), nil
}
