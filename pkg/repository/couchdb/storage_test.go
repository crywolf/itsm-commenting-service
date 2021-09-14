package couchdb_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment"
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment/listing"
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/entity"
	"github.com/KompiTech/itsm-commenting-service/pkg/event"
	"github.com/KompiTech/itsm-commenting-service/pkg/mocks"
	"github.com/KompiTech/itsm-commenting-service/testutils"
	"github.com/go-kivik/couchdb/v3/chttp"
	"github.com/go-kivik/kivik/v3/driver"
	"github.com/go-kivik/kivikmock/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAddComment(t *testing.T) {
	logger, _ := testutils.NewTestLogger()
	defer func() { _ = logger.Sync() }()

	channelID := "e27ddcd0-0e1f-4bc5-93df-f6f04155beec"
	orgID := "a897a407-e41b-4b14-924a-39f5d5a8038f"

	t.Run("with valid comment", func(t *testing.T) {
		validator := new(mocks.ValidatorMock)
		validator.On("Validate", mock.AnythingOfType("comment.Comment")).Return(nil)

		events := new(mocks.EventServiceMock)
		queue := new(mocks.QueueMock)
		events.On("NewQueue", event.UUID(channelID), event.UUID(orgID)).Return(queue, nil)
		queue.On("AddCreateEvent", mock.AnythingOfType("comment.Comment"), "comment").Return(nil)
		queue.On("PublishEvents").Return(nil)

		couchMock, s := testutils.NewCouchDBMock(context.Background(), logger, validator, events)

		db := couchMock.NewDB()
		couchMock.ExpectDB().WithName(testutils.DatabaseName(channelID, "comment")).WillReturn(db)
		db.ExpectPut()

		c := comment.Comment{
			Text:   "Test comment 1",
			Entity: entity.NewEntity("incident", "f49d5fd5-8da4-4779-b5ba-32e78aa2c444"),
			CreatedBy: &comment.UserInfo{
				UUID:           "7e0d38d1-e5f5-4211-b2aa-3b142e4da80e",
				Name:           "Andy",
				Surname:        "Orange",
				OrgName:        "a897a407-e41b-4b14-924a-39f5d5a8038f.kompitech.com",
				OrgDisplayName: "Kompitech",
			},
		}

		newC, err := s.AddComment(context.Background(), c, channelID, "comment")
		assert.Nil(t, err)
		assert.Equal(t, "38316161-3035-4864-ad30-6231392d3433", newC.UUID)

		validator.AssertExpectations(t)
		events.AssertExpectations(t)
		queue.AssertExpectations(t)
	})

	t.Run("with invalid comment", func(t *testing.T) {
		validator := new(mocks.ValidatorMock)
		validator.On("Validate", mock.AnythingOfType("comment.Comment")).Return(errors.New("invalid comment"))

		couchMock, s := testutils.NewCouchDBMock(context.Background(), logger, validator, nil)

		db := couchMock.NewDB()
		couchMock.ExpectDB().WithName(testutils.DatabaseName(channelID, "comment")).WillReturn(db)

		c := comment.Comment{
			Text:   "Test comment 1",
			Entity: entity.NewEntity("incident", "f49d5fd5-8da4-4779-b5ba-32e78aa2c444"),
		}

		newC, err := s.AddComment(context.Background(), c, channelID, "comment")
		assert.Error(t, err)
		assert.EqualErrorf(t, err, "invalid comment", "errors are not equal")

		assert.Nil(t, newC)

		validator.AssertNumberOfCalls(t, "Validate", 1)
	})

	t.Run("when comment with the same UUID already exists", func(t *testing.T) {
		validator := new(mocks.ValidatorMock)
		validator.On("Validate", mock.AnythingOfType("comment.Comment")).Return(nil)

		couchMock, s := testutils.NewCouchDBMock(context.Background(), logger, validator, nil)

		db := couchMock.NewDB()
		couchMock.ExpectDB().WithName(testutils.DatabaseName(channelID, "comment")).WillReturn(db)

		db.ExpectPut().WillReturnError(&chttp.HTTPError{
			Response: &http.Response{
				StatusCode: 409,
			},
		})

		c := comment.Comment{
			Text:   "Test comment 1",
			Entity: entity.NewEntity("incident", "f49d5fd5-8da4-4779-b5ba-32e78aa2c444"),
		}

		newC, err := s.AddComment(context.Background(), c, channelID, "comment")
		assert.Error(t, err)
		assert.EqualErrorf(t, err, "Comment could not be added: Comment already exists", "errors are not equal")
		assert.Nil(t, newC)

		validator.AssertNumberOfCalls(t, "Validate", 1)
	})
}

func TestGetComment(t *testing.T) {
	logger, _ := testutils.NewTestLogger()
	defer func() { _ = logger.Sync() }()

	channelID := "e27ddcd0-0e1f-4bc5-93df-f6f04155beec"

	t.Run("when comment does not exist", func(t *testing.T) {
		couchMock, s := testutils.NewCouchDBMock(context.Background(), logger, nil, nil)

		uuid := "cb2fe2a7-ab9f-4f6d-9fd6-c7c209403cf0"

		db := couchMock.NewDB()
		couchMock.ExpectDB().WithName(testutils.DatabaseName(channelID, "comment")).WillReturn(db)

		db.ExpectGet().WithDocID(uuid).WillExecute(func(ctx context.Context, arg0 string, options map[string]interface{}) (*driver.Document, error) {
			return &driver.Document{}, &chttp.HTTPError{
				Response: &http.Response{
					StatusCode: 404,
				},
			}
		})

		c, err := s.GetComment(context.Background(), uuid, channelID, "comment")
		assert.Error(t, err)
		assert.EqualErrorf(t, err, "Comment could not be retrieved: Comment with uuid='cb2fe2a7-ab9f-4f6d-9fd6-c7c209403cf0' does not exist", "errors are not equal")
		assert.Equal(t, comment.Comment{}, c)
	})

	t.Run("when comment exists", func(t *testing.T) {
		couchMock, s := testutils.NewCouchDBMock(context.Background(), logger, nil, nil)

		uuid := "cb2fe2a7-ab9f-4f6d-9fd6-c7c209403cf0"
		dbC := comment.Comment{
			UUID:   uuid,
			Entity: entity.NewEntity("incident", "f49d5fd5-8da4-4779-b5ba-32e78aa2c444"),
			Text:   "Some comment",
			CreatedBy: &comment.UserInfo{
				UUID:           "f49d5fd5-8da4-4779-b5ba-32e78aa2c444",
				Name:           "Joseph",
				Surname:        "Board",
				OrgName:        "a897a407-e41b-4b14-924a-39f5d5a8038f.kompitech.com",
				OrgDisplayName: "Kompitech",
			},
			CreatedAt: time.Now().Format(time.RFC3339),
		}

		db := couchMock.NewDB()
		couchMock.ExpectDB().WithName(testutils.DatabaseName(channelID, "comment")).WillReturn(db)
		row, err := kivikmock.Document(dbC)
		assert.Nil(t, err)
		db.ExpectGet().WithDocID(uuid).WillReturn(row)

		res, err := s.GetComment(context.Background(), uuid, channelID, "comment")
		assert.NoError(t, err)
		assert.Equal(t, dbC, res)
	})
}

func TestQueryComments(t *testing.T) {
	logger, _ := testutils.NewTestLogger()
	defer func() { _ = logger.Sync() }()

	channelID := "e27ddcd0-0e1f-4bc5-93df-f6f04155beec"

	t.Run("with invalid query", func(t *testing.T) {
		couchMock, s := testutils.NewCouchDBMock(context.Background(), logger, nil, nil)

		db := couchMock.NewDB()
		couchMock.ExpectDB().WithName(testutils.DatabaseName(channelID, "comment")).WillReturn(db)

		query := map[string]interface{}{"invalid query": ""}
		db.ExpectFind().WithQuery(query).WillReturnError(&chttp.HTTPError{
			Response: &http.Response{
				StatusCode: 400,
			},
			Reason: "invalid query",
		})

		res, err := s.QueryComments(context.Background(), query, channelID, "comment")
		assert.Error(t, err)
		assert.EqualErrorf(t, err, "invalid query", "errors are not equal")
		assert.Equal(t, listing.QueryResult{}, res)
	})

	t.Run("with valid query", func(t *testing.T) {
		couchMock, s := testutils.NewCouchDBMock(context.Background(), logger, nil, nil)

		db := couchMock.NewDB()
		couchMock.ExpectDB().WithName(testutils.DatabaseName(channelID, "comment")).WillReturn(db)

		query := map[string]interface{}{"valid": "query"}

		uuid := "cb2fe2a7-ab9f-4f6d-9fd6-c7c209403cf0"

		now := time.Now().Format(time.RFC3339)
		dbC := comment.Comment{
			UUID:   uuid,
			Entity: entity.NewEntity("incident", "f49d5fd5-8da4-4779-b5ba-32e78aa2c444"),
			Text:   "Some comment",
			CreatedBy: &comment.UserInfo{
				UUID:           "f49d5fd5-8da4-4779-b5ba-32e78aa2c444",
				Name:           "Joseph",
				Surname:        "Board",
				OrgName:        "a897a407-e41b-4b14-924a-39f5d5a8038f.kompitech.com",
				OrgDisplayName: "Kompitech",
			},
			CreatedAt: now,
		}

		cMap := map[string]interface{}{
			"uuid":   "cb2fe2a7-ab9f-4f6d-9fd6-c7c209403cf0",
			"entity": "incident:f49d5fd5-8da4-4779-b5ba-32e78aa2c444",
			"text":   "Some comment",
			"created_by": map[string]interface{}{
				"name":             "Joseph",
				"surname":          "Board",
				"uuid":             "f49d5fd5-8da4-4779-b5ba-32e78aa2c444",
				"org_name":         "a897a407-e41b-4b14-924a-39f5d5a8038f.kompitech.com",
				"org_display_name": "Kompitech",
			},
			"created_at": now,
		}

		doc, err := json.Marshal(dbC)
		require.NoError(t, err)
		db.ExpectFind().WithQuery(query).WillReturn(kivikmock.NewRows().
			AddRow(&driver.Row{ID: uuid, Doc: doc}).
			AddRow(&driver.Row{ID: uuid, Doc: doc}))

		res, err := s.QueryComments(context.Background(), query, channelID, "comment")
		assert.NoError(t, err)
		assert.NotEmpty(t, res)
		assert.Equal(t, listing.QueryResult{Result: []map[string]interface{}{cMap, cMap}}, res)
	})
}

func TestMarkAsReadByUser(t *testing.T) {
	logger, _ := testutils.NewTestLogger()
	defer func() { _ = logger.Sync() }()

	channelID := "e27ddcd0-0e1f-4bc5-93df-f6f04155beec"

	t.Run("when comment exists", func(t *testing.T) {
		validator := new(mocks.ValidatorMock)
		validator.On("Validate", mock.AnythingOfType("comment.Comment")).Return(nil)

		couchMock, s := testutils.NewCouchDBMock(context.Background(), logger, validator, nil)

		uuid := "cb2fe2a7-ab9f-4f6d-9fd6-c7c209403cf0"

		dbC := comment.Comment{
			UUID:   uuid,
			Entity: entity.NewEntity("incident", "f49d5fd5-8da4-4779-b5ba-32e78aa2c444"),
			Text:   "Some comment",
			CreatedBy: &comment.UserInfo{
				UUID:           "f49d5fd5-8da4-4779-b5ba-32e78aa2c444",
				Name:           "Joseph",
				Surname:        "Board",
				OrgName:        "a897a407-e41b-4b14-924a-39f5d5a8038f.kompitech.com",
				OrgDisplayName: "Kompitech",
			},
			CreatedAt: time.Now().Format(time.RFC3339),
		}

		db := couchMock.NewDB()
		couchMock.ExpectDB().WithName(testutils.DatabaseName(channelID, "comment")).WillReturn(db)
		row, err := kivikmock.Document(dbC)
		assert.Nil(t, err)
		db.ExpectGet().WithDocID(uuid).WillReturn(row)

		readBy := comment.ReadBy{
			Time: time.Now().Format(time.RFC3339),
			User: comment.UserInfo{
				UUID:           "439e2d19-8d50-405d-ad8e-cd33df344086",
				Name:           "Joe",
				Surname:        "Potato",
				OrgName:        "a897a407-e41b-4b14-924a-39f5d5a8038f.kompitech.com",
				OrgDisplayName: "Kompitech",
			},
		}

		db.ExpectPut().WithDocID(uuid)

		res, err := s.MarkAsReadByUser(context.Background(), uuid, readBy, channelID, "comment")
		assert.NoError(t, err)
		assert.False(t, res)
	})

	t.Run("when comment does not exist", func(t *testing.T) {
		couchMock, s := testutils.NewCouchDBMock(context.Background(), logger, nil, nil)

		uuid := "cb2fe2a7-ab9f-4f6d-9fd6-c7c209403cf0"

		db := couchMock.NewDB()
		couchMock.ExpectDB().WithName(testutils.DatabaseName(channelID, "comment")).WillReturn(db)

		db.ExpectGet().WithDocID(uuid).WillExecute(func(ctx context.Context, arg0 string, options map[string]interface{}) (*driver.Document, error) {
			return &driver.Document{}, &chttp.HTTPError{
				Response: &http.Response{
					StatusCode: 404,
				},
			}
		})

		readBy := comment.ReadBy{
			Time: time.Now().Format(time.RFC3339),
			User: comment.UserInfo{
				UUID:           "439e2d19-8d50-405d-ad8e-cd33df344086",
				Name:           "Joe",
				Surname:        "Potato",
				OrgName:        "a897a407-e41b-4b14-924a-39f5d5a8038f.kompitech.com",
				OrgDisplayName: "Kompitech",
			},
		}

		res, err := s.MarkAsReadByUser(context.Background(), uuid, readBy, channelID, "comment")
		assert.Error(t, err)
		assert.EqualErrorf(t, err, "Comment with uuid='cb2fe2a7-ab9f-4f6d-9fd6-c7c209403cf0' does not exist", "errors are not equal")
		assert.False(t, res)
	})

	t.Run("when comment was already read by user", func(t *testing.T) {
		couchMock, s := testutils.NewCouchDBMock(context.Background(), logger, nil, nil)

		commentUUID := "cb2fe2a7-ab9f-4f6d-9fd6-c7c209403cf0"
		userUUID := "f49d5fd5-8da4-4779-b5ba-32e78aa2c444"

		userInfo := comment.UserInfo{
			UUID:           userUUID,
			Name:           "Joseph",
			Surname:        "Board",
			OrgName:        "a897a407-e41b-4b14-924a-39f5d5a8038f.kompitech.com",
			OrgDisplayName: "Kompitech",
		}

		dbC := comment.Comment{
			UUID:   commentUUID,
			Entity: entity.NewEntity("incident", "f49d5fd5-8da4-4779-b5ba-32e78aa2c444"),
			Text:   "Some comment",
			CreatedBy: &comment.UserInfo{
				UUID:           "439e2d19-8d50-405d-ad8e-cd33df344086",
				Name:           "Joseph",
				Surname:        "Board",
				OrgName:        "a897a407-e41b-4b14-924a-39f5d5a8038f.kompitech.com",
				OrgDisplayName: "Kompitech",
			},
			ReadBy: comment.ReadByList{
				{
					User: userInfo,
					Time: time.Now().Format(time.RFC3339),
				},
			},
			CreatedAt: time.Now().Format(time.RFC3339),
		}

		db := couchMock.NewDB()
		couchMock.ExpectDB().WithName(testutils.DatabaseName(channelID, "comment")).WillReturn(db)
		row, err := kivikmock.Document(dbC)
		assert.Nil(t, err)
		db.ExpectGet().WithDocID(commentUUID).WillReturn(row)

		readBy := comment.ReadBy{
			Time: time.Now().Format(time.RFC3339),
			User: userInfo,
		}

		res, err := s.MarkAsReadByUser(context.Background(), commentUUID, readBy, channelID, "comment")
		assert.NoError(t, err)
		assert.True(t, res)
	})
}

func TestCreateDatabase(t *testing.T) {
	logger, _ := testutils.NewTestLogger()
	defer func() { _ = logger.Sync() }()

	channelID := "e27ddcd0-0e1f-4bc5-93df-f6f04155beec"

	t.Run("when database already exists", func(t *testing.T) {
		couchMock, s := testutils.NewCouchDBMock(context.Background(), logger, nil, nil)

		couchMock.ExpectDBExists().WithName(testutils.DatabaseName(channelID, "comment")).WillReturn(true)

		existed, err := s.CreateDatabase(context.Background(), channelID, "comment")
		assert.Nil(t, err)
		assert.Equal(t, true, existed)
	})

	t.Run("when database does not exist", func(t *testing.T) {
		couchMock, s := testutils.NewCouchDBMock(context.Background(), logger, nil, nil)

		couchMock.ExpectDBExists().WithName(testutils.DatabaseName(channelID, "comment")).WillReturn(false)
		couchMock.ExpectCreateDB().WithName(testutils.DatabaseName(channelID, "comment"))
		db := couchMock.NewDB()
		couchMock.ExpectDB().WithName(testutils.DatabaseName(channelID, "comment")).WillReturn(db)
		db.ExpectCreateIndex()

		existed, err := s.CreateDatabase(context.Background(), channelID, "comment")
		assert.Nil(t, err)
		assert.Equal(t, false, existed)
	})
}
