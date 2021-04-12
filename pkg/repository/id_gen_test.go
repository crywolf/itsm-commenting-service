package repository

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateID(t *testing.T) {
	id, err := GenerateUUID(nil)
	require.NoError(t, err)
	require.Len(t, id, 36)
	uuidRe := regexp.MustCompile(`(?i)^[0-9A-F]{8}-[0-9A-F]{4}-4[0-9A-F]{3}-[89AB][0-9A-F]{3}-[0-9A-F]{12}$`)
	require.Regexp(t, uuidRe, id)
}
