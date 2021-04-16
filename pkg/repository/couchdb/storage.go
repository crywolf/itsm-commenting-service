package couchdb

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment"
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment/listing"
	"github.com/KompiTech/itsm-commenting-service/pkg/repository"
	"go.uber.org/zap"

	_ "github.com/go-kivik/couchdb/v3" // The CouchDB driver
	"github.com/go-kivik/couchdb/v3/chttp"
	"github.com/go-kivik/kivik/v3"
)

const (
	// database names
	dbComments  = "comments"
	dbWorknotes = "worknotes"

	// defaultPageSize is the default couchDB value for 'limit' in 'find' query
	defaultPageSize = 25
)

// DBStorage storage stores data in couchdb
type DBStorage struct {
	client   *kivik.Client
	logger   *zap.Logger
	rand     io.Reader
	username string
	passwd   string
}

// Config contains values for the data source
type Config struct {
	Client   *kivik.Client
	Rand     io.Reader
	Host     string
	Port     string
	Username string
	Passwd   string
}

// NewStorage creates new couchdb storage with initialized client
// Also creates nonexistent databases
func NewStorage(logger *zap.Logger, cfg Config) *DBStorage {
	var client *kivik.Client
	var err error

	if cfg.Client == nil {
		client, err = kivik.New(
			"couch",
			fmt.Sprintf("http://%s:%s@%s:%s/", cfg.Username, cfg.Passwd, cfg.Host, cfg.Port),
		)
		if err != nil {
			logger.Fatal("couchdb client initialization failed", zap.Error(err))
		}
	} else {
		client = cfg.Client
	}

	ctx := context.TODO()

	commExists, err := client.DBExists(ctx, dbComments)
	if err != nil {
		logger.Fatal("couchdb connection failed", zap.Error(err))
	}

	if !commExists {
		err := client.CreateDB(ctx, dbComments)
		if err != nil {
			logger.Fatal("couchdb database creation failed", zap.Error(err))
		}

		db := client.DB(ctx, dbComments)

		index := map[string]interface{}{
			"fields": []string{"created_at"},
		}
		err = db.CreateIndex(ctx, "", "", index)
		if err != nil {
			logger.Fatal("couchdb database index creation failed", zap.Error(err))
		}
	}

	wnExists, err := client.DBExists(ctx, dbWorknotes)
	if err != nil {
		logger.Fatal("couchdb connection failed", zap.Error(err))
	}

	if !wnExists {
		err := client.CreateDB(ctx, dbWorknotes)
		if err != nil {
			logger.Fatal("couchdb database creation failed", zap.Error(err))
		}

		db := client.DB(ctx, dbWorknotes)

		index := map[string]interface{}{
			"fields": []string{"created_at"},
		}
		err = db.CreateIndex(ctx, "", "", index)
		if err != nil {
			logger.Fatal("couchdb database index creation failed", zap.Error(err))
		}
	}

	return &DBStorage{
		client: client,
		logger: logger,
		rand:   cfg.Rand,
	}
}

// AddComment saves the given comment to the database and returns it's ID
func (s *DBStorage) AddComment(c comment.Comment) (string, error) {
	dbName := dbComments
	ctx := context.TODO()

	db := s.client.DB(ctx, dbName)

	uuid, err := repository.GenerateUUID(s.rand)
	if err != nil {
		return "", err
	}

	newC := Comment{
		UUID:      uuid,
		Entity:    c.Entity,
		Text:      c.Text,
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	rev, err := db.Put(ctx, uuid, newC)
	if err != nil {
		s.logger.Warn("CouchDB PUT failed", zap.Error(err))

		var httpError *chttp.HTTPError
		if errors.As(err, &httpError) {
			reason := httpError.Reason

			if httpError.StatusCode() == http.StatusConflict {
				reason = "Comment already exists"
			}

			eMsg := fmt.Sprintf("Comment could not be added: %s", reason)
			return "", ErrorConflict(eMsg)
		}

		return "", err
	}

	s.logger.Info(fmt.Sprintf("Comment inserted with revision %s", rev))

	return uuid, nil
}

// GetComment returns comment with the specified ID
func (s *DBStorage) GetComment(id string) (comment.Comment, error) {
	ctx := context.TODO()

	var c comment.Comment

	db := s.client.DB(ctx, dbComments)

	row := db.Get(ctx, id)
	err := row.ScanDoc(&c)
	if err != nil {
		s.logger.Warn("CouchDB GET failed", zap.Error(err))

		var httpError *chttp.HTTPError
		if errors.As(err, &httpError) {
			reason := httpError.Reason

			if httpError.StatusCode() == http.StatusNotFound {
				reason = fmt.Sprintf("Comment with uuid='%s' does not exist", id)
			}

			eMsg := fmt.Sprintf("Comment could not be retrieved: %s", reason)
			return c, ErrorNorFound(eMsg)
		}

		return c, err
	}

	s.logger.Info(fmt.Sprintf("Comment fetched %s", c))

	return c, nil
}

// QueryComments finds documents using a declarative JSON querying syntax
func (s *DBStorage) QueryComments(query map[string]interface{}) (listing.QueryResult, error) {
	ctx := context.TODO()

	var docs []map[string]interface{}

	db := s.client.DB(ctx, dbComments)

	rows, err := db.Find(ctx, query)
	if err != nil {
		s.logger.Warn("CouchDB FIND failed", zap.Error(err))

		var httpError *chttp.HTTPError
		if errors.As(err, &httpError) {
			if httpError.StatusCode() == http.StatusBadRequest {
				return listing.QueryResult{}, ErrorBadRequest(httpError.Reason)
			}
		}

		return listing.QueryResult{}, err
	}

	for rows.Next() {
		var doc map[string]interface{}
		err := rows.ScanDoc(&doc)
		if err != nil {
			return listing.QueryResult{}, err
		}

		docs = append(docs, doc)
	}

	var bookmark string
	limit := defaultPageSize

	// do not set bookmark if the length of returned results is smaller
	// then the requested limit
	lim, exists := query["limit"]
	if exists {
		fLimit, ok := lim.(float64)
		if ok {
			limit = int(fLimit)
		}
	}

	if len(docs) == limit {
		bookmark = rows.Bookmark()
	}

	result := listing.QueryResult{
		Result:   docs,
		Bookmark: bookmark,
	}

	return result, err
}
