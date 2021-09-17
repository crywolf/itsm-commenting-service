package couchdb

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment"
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment/listing"
	"github.com/KompiTech/itsm-commenting-service/pkg/event"
	"github.com/KompiTech/itsm-commenting-service/pkg/repository"
	"github.com/go-kivik/couchdb/v3" // The CouchDB driver
	"github.com/go-kivik/couchdb/v3/chttp"
	"github.com/go-kivik/kivik/v3"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
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
	events    event.Service
}

// Config contains values for the data source
type Config struct {
	Client       *kivik.Client
	CaPath       string
	Rand         io.Reader
	Validator    Validator
	EventService event.Service
	Host         string
	Port         string
	Username     string
	Passwd       string
}

// NewStorage creates new couchdb storage with initialized client
func NewStorage(ctx context.Context, logger *zap.Logger, cfg Config) *DBStorage {
	var client *kivik.Client
	var err error
	var caBytes []byte

	tlsOn := cfg.CaPath != ""

	if cfg.Client == nil {
		schema := "http"
		if tlsOn {
			schema = "https"
		}

		client, err = kivik.New(
			"couch",
			fmt.Sprintf("%s://%s:%s/", schema, cfg.Host, cfg.Port),
		)
		if err != nil {
			logger.Fatal("couchdb client initialization failed", zap.Error(err))
		}

		if tlsOn {
			caBytes, err = ioutil.ReadFile(cfg.CaPath)
			if err != nil {
				logger.Fatal("couchdb client initialization failed - certificate", zap.Error(err))
			}

			certPool := x509.NewCertPool()
			certPool.AppendCertsFromPEM(caBytes)

			couchAuthenticator := couchdb.SetTransport(&http.Transport{
				TLSClientConfig: &tls.Config{
					RootCAs: certPool,
				},
			})
			if err := client.Authenticate(ctx, couchAuthenticator); err != nil {
				logger.Fatal("couchdb authentication failed", zap.Error(err))
			}
		}

		if err := client.Authenticate(ctx, &chttp.CookieAuth{Username: cfg.Username, Password: cfg.Passwd}); err != nil {
			logger.Fatal("couchdb authentication failed", zap.Error(err))
		}

		waitForCouchDB(logger, client, tlsOn)
	} else {
		client = cfg.Client
	}

	return &DBStorage{
		client:    client,
		logger:    logger,
		rand:      cfg.Rand,
		validator: cfg.Validator,
		events:    cfg.EventService,
	}
}

// Client returns kivik client connection handle
func (s *DBStorage) Client() *kivik.Client {
	return s.client
}

// AddComment saves the given comment to the database and returns it's ID
func (s *DBStorage) AddComment(ctx context.Context, c comment.Comment, channelID, assetType string) (*comment.Comment, error) {
	dbName := databaseName(channelID, assetType)

	db := s.client.DB(ctx, dbName)

	uuid, err := repository.GenerateUUID(s.rand)
	if err != nil {
		s.logger.Error("could not generate UUID", zap.Error(err))
		return nil, err
	}

	c.UUID = uuid
	c.CreatedAt = time.Now().Format(time.RFC3339)

	err = s.validator.Validate(c)
	if err != nil {
		s.logger.Error("invalid "+assetType, zap.Error(err))
		return nil, err
	}

	rev, err := db.Put(ctx, uuid, c)
	if err != nil {
		s.logger.Warn("CouchDB PUT failed", zap.Error(err))

		var httpError *chttp.HTTPError
		if errors.As(err, &httpError) {
			if httpError.StatusCode() == http.StatusConflict {
				reason := fmt.Sprintf("%s already exists", strings.Title(assetType))
				eMsg := fmt.Sprintf("%s could not be added: %s", strings.Title(assetType), reason)
				return nil, ErrorConflict(eMsg)
			}

			eMsg := fmt.Sprintf("%s could not be added: %s", strings.Title(assetType), httpError.Reason)
			return nil, repository.NewError(eMsg, http.StatusInternalServerError)
		}

		return nil, err
	}

	s.logger.Info(fmt.Sprintf("%s inserted with revision %s", strings.Title(assetType), rev))

	q, err := s.events.NewQueue(event.UUID(channelID), event.UUID(c.CreatedBy.OrgID()))
	if err != nil {
		msg := "could not create event queue"
		s.logger.Error(msg, zap.Error(err))
		s.rollback(ctx, db, uuid, rev, assetType)

		return nil, fmt.Errorf("%s: %v", msg, err)
	}

	if err = q.AddCreateEvent(c, assetType); err != nil {
		msg := "could not create event"
		s.logger.Error(msg, zap.Error(err))
		s.rollback(ctx, db, uuid, rev, assetType)

		return nil, fmt.Errorf("%s: %v", msg, err)
	}

	if err = q.PublishEvents(); err != nil {
		msg := "could not publish events"
		s.logger.Error(msg, zap.Error(err))
		s.rollback(ctx, db, uuid, rev, assetType)

		return nil, fmt.Errorf("%s: %v", msg, err)
	}

	return &c, nil
}

func (s *DBStorage) rollback(ctx context.Context, db *kivik.DB, uuid string, rev string, assetType string) {
	_, err := db.Delete(ctx, uuid, rev)
	if err != nil {
		s.logger.Error(fmt.Sprintf("could not delete %s:%s (rollback)", assetType, uuid), zap.Error(err))
		return
	}

	s.logger.Info(fmt.Sprintf("%s:%s deleted (rollback)", assetType, uuid))
}

// GetComment returns comment with the specified ID
func (s *DBStorage) GetComment(ctx context.Context, id, channelID, assetType string) (comment.Comment, error) {
	dbName := databaseName(channelID, assetType)

	var c comment.Comment

	db := s.client.DB(ctx, dbName)

	row := db.Get(ctx, id)
	err := row.ScanDoc(&c)
	if err != nil {
		s.logger.Warn("CouchDB GET failed", zap.Error(err))

		var httpError *chttp.HTTPError
		if errors.As(err, &httpError) {
			if httpError.StatusCode() == http.StatusNotFound {
				reason := fmt.Sprintf("%s with uuid='%s' does not exist", strings.Title(assetType), id)
				eMsg := fmt.Sprintf("%s could not be retrieved: %s", strings.Title(assetType), reason)
				return c, ErrorNorFound(eMsg)
			}

			eMsg := fmt.Sprintf("%s could not be retrieved: %s", strings.Title(assetType), httpError.Reason)
			return c, repository.NewError(eMsg, http.StatusInternalServerError)
		}

		return c, err
	}

	s.logger.Info(fmt.Sprintf("%s fetched %v", strings.Title(assetType), c))

	return c, nil
}

// QueryComments finds documents using a declarative JSON querying syntax
func (s *DBStorage) QueryComments(ctx context.Context, query map[string]interface{}, channelID, assetType string) (listing.QueryResult, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "itsm-commenting-service-query-dbstorage")
	defer span.Finish()
	dbName := databaseName(channelID, assetType)

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
func (s *DBStorage) MarkAsReadByUser(ctx context.Context, id string, readBy comment.ReadBy, channelID, assetType string) (bool, error) {
	dbName := databaseName(channelID, assetType)

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
				reason = fmt.Sprintf("%s with uuid='%s' does not exist", strings.Title(assetType), id)
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
				reason = fmt.Sprintf("%s could not be mark as read", strings.Title(assetType))
			}

			return false, ErrorConflict(reason)
		}

		return false, err
	}

	s.logger.Info(fmt.Sprintf("%s updated %#v", strings.Title(assetType), c))

	return false, nil
}

// CreateDatabase creates new DB if it does not exist. It returns true if database already existed.
func (s *DBStorage) CreateDatabase(ctx context.Context, channelID, assetType string) (bool, error) {
	dbName := databaseName(channelID, assetType)

	dbExists, err := s.client.DBExists(ctx, dbName)
	if err != nil {
		s.logger.Error("couchdb connection failed", zap.Error(err))
		return false, err
	}

	if dbExists {
		return true, nil
	}

	err = s.client.CreateDB(ctx, dbName)
	if err != nil {
		s.logger.Error("couchdb database creation failed", zap.Error(err))
		return false, err
	}

	db := s.client.DB(ctx, dbName)

	index := map[string]interface{}{
		"fields": []string{"created_at"},
	}
	err = db.CreateIndex(ctx, "", "", index)
	if err != nil {
		s.logger.Error("couchdb database index creation failed", zap.Error(err))
		return false, err
	}

	return false, nil
}

func databaseName(channelID, assetType string) string {
	return fmt.Sprintf("p_%s_%s", channelID, pluralize(assetType))
}

func pluralize(assetType string) string {
	return fmt.Sprintf("%ss", assetType)
}

// waitForCouchDB repeatedly tries to ping DB server until it is ready for requests or timeout expires
func waitForCouchDB(logger *zap.Logger, client *kivik.Client, tlsOn bool) {
	logger.Info("Waiting for CouchDB to become ready...")
	maxIters := 100 // default 100 * 100ms = 10 seconds
	iter := 0
	stepMs := 100
	reportEveryIter := 10
	step := time.Duration(stepMs) * time.Millisecond

	// wait for couchdb to response to ping
	for {
		on, err := client.Ping(context.TODO())
		if err == nil && on {
			break
		}
		time.Sleep(step)
		iter++
		if iter == maxIters {
			logger.Fatal("Waited for CouchDB for too long. Something is wrong. If this happens while running tests check docker and docker-compose status and try again.")
		} else if (iter > 0) && (iter%reportEveryIter == 0) {
			logger.Warn("CouchDB still not ready, waiting...", zap.Int64("ms", int64(iter)*int64(stepMs)))
		}
	}
	logger.Info("CouchDB became ready in", zap.Int64("ms", int64(iter)*int64(stepMs)), zap.Bool("TLS mode", tlsOn))
}
