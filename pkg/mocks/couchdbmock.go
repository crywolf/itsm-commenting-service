package mocks

import (
	"context"
	"strings"

	"github.com/KompiTech/itsm-commenting-service/pkg/event"
	"github.com/KompiTech/itsm-commenting-service/pkg/repository/couchdb"
	"github.com/go-kivik/kivikmock/v3"
	"go.uber.org/zap"
)

// GeneratedCommentUUID is UUID of the newly created comment/worknote generated by pseudorandom generator
var GeneratedCommentUUID = "38316161-3035-4864-ad30-6231392d3433"

// NewCouchDBMock creates new CouchDB mock
func NewCouchDBMock(ctx context.Context, logger *zap.Logger, validator couchdb.Validator, events event.Service) (*kivikmock.Client, *couchdb.DBStorage) {
	client, mock, err := kivikmock.New()
	if err != nil {
		panic(err)
	}

	// rand is used as deterministic UUID generator
	// repository.GenerateUID(rand) returns always "38316161-3035-4864-ad30-6231392d3433"
	rand := strings.NewReader("81aa058d-0b19-43e9-82ae-a7bca2457f10") // pseudo-random seed

	storage := couchdb.NewStorage(ctx, logger, couchdb.Config{
		Client:       client,
		Rand:         rand,
		Validator:    validator,
		EventService: events,
	})

	return mock, storage
}
