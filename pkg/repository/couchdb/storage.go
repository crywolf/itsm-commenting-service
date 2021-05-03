package couchdb

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment"
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment/listing"
	"github.com/KompiTech/itsm-commenting-service/pkg/repository"
	"go.uber.org/zap"

	_ "github.com/go-kivik/couchdb/v3" // The CouchDB driver
	"github.com/go-kivik/couchdb/v3/chttp"
	"github.com/go-kivik/kivik/v3"
)

const (
	// defaultPageSize is the default couchDB value for 'limit' in 'find' query
	defaultPageSize = 25
)

// DBStorage storage stores data in couchdb
type DBStorage struct {
	client    *kivik.Client
	logger    *zap.Logger
	rand      io.Reader
	validator Validator
	username  string
	passwd    string
}

// Config contains values for the data source
type Config struct {
	Client    *kivik.Client
	Rand      io.Reader
	Validator Validator
	Host      string
	Port      string
	Username  string
	Passwd    string
}

// NewStorage creates new couchdb storage with initialized client
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

	return &DBStorage{
		client:    client,
		logger:    logger,
		rand:      cfg.Rand,
		validator: cfg.Validator,
	}
}

// AddComment saves the given comment to the database and returns it's ID
func (s *DBStorage) AddComment(c comment.Comment, channelID, assetType string) (string, error) {
	dbName := databaseName(channelID, assetType)
	ctx := context.TODO()

	db := s.client.DB(ctx, dbName)

	uuid, err := repository.GenerateUUID(s.rand)
	if err != nil {
		s.logger.Error("could not generate UUID", zap.Error(err))
		return "", err
	}

	c.UUID = uuid
	err = s.validator.Validate(c)
	if err != nil {
		s.logger.Error("invalid comment", zap.Error(err))
		return "", err
	}

	rev, err := db.Put(ctx, uuid, c)
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
func (s *DBStorage) GetComment(id, channelID, assetType string) (comment.Comment, error) {
	dbName := databaseName(channelID, assetType)
	ctx := context.TODO()

	var c comment.Comment

	db := s.client.DB(ctx, dbName)

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

	s.logger.Info(fmt.Sprintf("Comment fetched %v", c))

	return c, nil
}

// QueryComments finds documents using a declarative JSON querying syntax
func (s *DBStorage) QueryComments(query map[string]interface{}, channelID, assetType string) (listing.QueryResult, error) {
	dbName := databaseName(channelID, assetType)
	ctx := context.TODO()

	var docs []map[string]interface{}
	docs = make([]map[string]interface{}, 0)

	db := s.client.DB(ctx, dbName)

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

// MarkAsReadByUser adds user info to read_by array in the comment with specified ID.
// It returns true if comment was already marked before to notify that resource was not changed.
func (s *DBStorage) MarkAsReadByUser(id string, readBy comment.ReadBy, channelID, assetType string) (bool, error) {
	dbName := databaseName(channelID, assetType)
	ctx := context.TODO()

	var c comment.Comment

	db := s.client.DB(ctx, dbName)

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

			return false, ErrorNorFound(reason)
		}

		return false, err
	}

	currentUserID := readBy.User.UUID

	for _, rb := range c.ReadBy {
		if rb.User.UUID == currentUserID {
			// comment was already read by user in the past
			return true, nil
		}
	}

	c.ReadBy = append(c.ReadBy, readBy)

	err = s.validator.Validate(c)
	if err != nil {
		s.logger.Error("invalid comment", zap.Error(err))
		return false, err
	}

	// updated comment with revision ID
	var uc struct {
		Rev string `json:"_rev"`
		comment.Comment
	}

	uc.Comment = c
	uc.Rev = row.Rev

	_, err = db.Put(ctx, uc.UUID, uc)
	if err != nil {
		s.logger.Warn("CouchDB PUT failed", zap.Error(err))

		var httpError *chttp.HTTPError
		if errors.As(err, &httpError) {
			reason := httpError.Reason

			if httpError.StatusCode() == http.StatusConflict {
				reason = "Comment could not be mark as read"
			}

			return false, ErrorConflict(reason)
		}

		return false, err
	}

	s.logger.Info(fmt.Sprintf("Comment updated %#v", c))

	return false, nil
}

// CreateDatabase creates new DB if it does not exist. It returns true if database already existed.
func (s *DBStorage) CreateDatabase(channelID, assetType string) (bool, error) {
	dbName := databaseName(channelID, assetType)
	ctx := context.TODO()

	dbExists, err := s.client.DBExists(ctx, dbName)
	if err != nil {
		s.logger.Fatal("couchdb connection failed", zap.Error(err))
		return false, err
	}

	if dbExists {
		return true, nil
	}

	err = s.client.CreateDB(ctx, dbName)
	if err != nil {
		s.logger.Fatal("couchdb database creation failed", zap.Error(err))
		return false, err
	}

	db := s.client.DB(ctx, dbName)

	index := map[string]interface{}{
		"fields": []string{"created_at"},
	}
	err = db.CreateIndex(ctx, "", "", index)
	if err != nil {
		s.logger.Fatal("couchdb database index creation failed", zap.Error(err))
		return false, err
	}

	return false, nil
}

func databaseName(channelID, assetType string) string {
	return fmt.Sprintf("%s_%s", channelID, pluralize(assetType))
}

func pluralize(assetType string) string {
	return fmt.Sprintf("%ss", assetType)
}
