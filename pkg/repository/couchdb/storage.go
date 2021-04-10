package couchdb

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment"
	"github.com/KompiTech/itsm-commenting-service/pkg/repository"
	"go.uber.org/zap"

	_ "github.com/go-kivik/couchdb" // The CouchDB driver
	"github.com/go-kivik/couchdb/chttp"
	"github.com/go-kivik/kivik"
)

const (
	dbComments  = "comments"
	dbWorknotes = "worknotes"
)

// DBStorage storage stores data in couchdb
type DBStorage struct {
	client   *kivik.Client
	logger   *zap.Logger
	username string
	passwd   string
}

// Config contains values for the data source
type Config struct {
	Host     string
	Port     string
	Username string
	Passwd   string
}

// NewStorage creates new couchdb storage with initialized client
// Also creates nonexistent databases
func NewStorage(logger *zap.Logger, cfg Config) *DBStorage {
	client, err := kivik.New(
		"couch",
		fmt.Sprintf("http://%s:%s@%s:%s/", cfg.Username, cfg.Passwd, cfg.Host, cfg.Port),
	)
	if err != nil {
		panic(err)
	}

	ctx := context.TODO()

	commExists, err := client.DBExists(ctx, dbComments)
	if err != nil {
		panic(err)
	}

	if !commExists {
		err := client.CreateDB(ctx, dbComments)
		if err != nil {
			panic(err)
		}
	}

	wnExists, err := client.DBExists(ctx, dbWorknotes)
	if err != nil {
		panic(err)
	}

	if !wnExists {
		err := client.CreateDB(ctx, dbWorknotes)
		if err != nil {
			panic(err)
		}
	}

	return &DBStorage{
		client: client,
		logger: logger,
	}
}

// AddComment saves the given asset to the database and returns it's ID
func (s *DBStorage) AddComment(c comment.Comment) (string, error) {
	dbName := dbComments
	ctx := context.TODO()

	db := s.client.DB(ctx, dbName)

	uuid, err := repository.GenerateUUID()
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
		s.logger.Error("CouchDB PUT failed", zap.Error(err))

		var httpError *chttp.HTTPError
		if errors.As(err, &httpError) {
			reason := httpError.Reason

			if httpError.StatusCode() == http.StatusConflict {
				reason = "Comment already exists"
			}

			msg := fmt.Sprintf("Comment could not be added: %s", reason)
			return "", repository.New(msg, httpError.StatusCode())
		}

		return "", err
	}

	s.logger.Info(fmt.Sprintf("Comment inserted with revision %s", rev))

	return uuid, nil
}

// GetComment returns a comment with the specified ID
func (s *DBStorage) GetComment(id string) (comment.Comment, error) {
	ctx := context.TODO()

	var c comment.Comment

	db := s.client.DB(ctx, dbComments)

	row := db.Get(ctx, id)
	err := row.ScanDoc(&c)
	if err != nil {
		s.logger.Error("CouchDB GET failed", zap.Error(err))

		var httpError *chttp.HTTPError
		if errors.As(err, &httpError) {
			reason := httpError.Reason

			if httpError.StatusCode() == http.StatusNotFound {
				reason = fmt.Sprintf("Comment with uuid='%s' does not exist", id)
			}

			msg := fmt.Sprintf("Comment could not be retrieved: %s", reason)
			return c, repository.New(msg, httpError.StatusCode())
		}

		return c, err
	}

	s.logger.Info(fmt.Sprintf("Comment fetched %s", c))

	return c, nil
}
