package repository

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// GenerateID returns a random UUID
func GenerateID() (string, error) {
	uuid, err := generateUUID()
	if err != nil {
		return "", errors.Wrap(err, "Could not generate UUID")
	}

	return uuid, nil
}

func generateUUID() (string, error) {
	rand.Seed(time.Now().UnixNano())

	b := make([]byte, 16)

	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return strings.ToLower(fmt.Sprintf("%X-%X-%X-%X-%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])), nil
}
