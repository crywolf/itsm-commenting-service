package testutils

import (
	"strings"

	"github.com/KompiTech/itsm-commenting-service/pkg/repository/couchdb"
	"github.com/go-kivik/kivikmock/v3"
	"go.uber.org/zap"
)

// NewCouchDBMock creates new CouchDB mock
func NewCouchDBMock(logger *zap.Logger) (*kivikmock.Client, *couchdb.DBStorage) {
	client, mock, err := kivikmock.New()
	if err != nil {
		panic(err)
	}

	// rand is used as deterministic UUID generator
	// repository.GenerateUID(rand) returns always "38316161-3035-4864-ad30-6231392d3433"
	rand := strings.NewReader("81aa058d-0b19-43e9-82ae-a7bca2457f10") // pseudo-random seed

	storage := couchdb.NewStorage(logger, couchdb.Config{
		Client: client,
		Rand:   rand,
	})

	return mock, storage
}
