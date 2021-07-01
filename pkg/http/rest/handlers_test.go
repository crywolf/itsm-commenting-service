package rest_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment"
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment/adding"
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment/listing"
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/entity"
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/user"
	"github.com/KompiTech/itsm-commenting-service/pkg/event"
	"github.com/KompiTech/itsm-commenting-service/pkg/http/rest"
	"github.com/KompiTech/itsm-commenting-service/pkg/http/rest/auth"
	"github.com/KompiTech/itsm-commenting-service/pkg/http/rest/validation"
	"github.com/KompiTech/itsm-commenting-service/pkg/mocks"
	"github.com/KompiTech/itsm-commenting-service/pkg/repository/memory"
	"github.com/KompiTech/itsm-commenting-service/testutils"
	"github.com/go-kivik/kivikmock/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

/*
  Some other tests that are not much important.
  Everything is tested in files with tests for specific handlers.
  This was just the first stage without service mocking, just to test the idea :)
  This file can be safely removed.
*/

func TestAddCommentDBMock(t *testing.T) {
	logger, _ := testutils.NewTestLogger()
	defer func() { _ = logger.Sync() }()

	bearerToken := "some valid Bearer token"
	channelID := "e27ddcd0-0e1f-4bc5-93df-f6f04155beec"
	orgID := "a897a407-e41b-4b14-924a-39f5d5a8038f"
	assetType := "comment"

	validator := new(mocks.ValidatorMock)
	validator.On("Validate", mock.AnythingOfType("comment.Comment")).Return(nil)

	events := new(mocks.EventServiceMock)
	queue := new(mocks.QueueMock)
	events.On("NewQueue", event.UUID(channelID), event.UUID(orgID)).Return(queue, nil)
	queue.On("AddCreateEvent", mock.AnythingOfType("comment.Comment"), assetType).Return(nil)
	queue.On("PublishEvents").Return(nil)

	couchMock, s := testutils.NewCouchDBMock(logger, validator, events)

	db := couchMock.NewDB()
	couchMock.ExpectDB().WithName(testutils.DatabaseName(channelID, assetType)).WillReturn(db)
	db.ExpectPut()

	as := new(mocks.AuthServiceMock)
	as.On("Enforce", assetType, auth.CreateAction, channelID, bearerToken).
		Return(true, nil)

	us := new(mocks.UserServiceMock)
	mockUserData := user.BasicInfo{
		UUID:           "2af4f493-0bd5-4513-b440-6cbb465feadb",
		Name:           "Alice",
		Surname:        "Cooper",
		OrgName:        "a897a407-e41b-4b14-924a-39f5d5a8038f.kompitech.com",
		OrgDisplayName: "Kompitech",
	}
	us.On("UserBasicInfo", mock.AnythingOfType("*http.Request")).
		Return(mockUserData, nil)

	adder := adding.NewService(s)

	pv, err := validation.NewPayloadValidator()
	require.NoError(t, err)

	server := rest.NewServer(rest.Config{
		Addr:             "service.url",
		AuthService:      as,
		UserService:      us,
		Logger:           logger,
		AddingService:    adder,
		PayloadValidator: pv,
	})

	payload := []byte(`{
		"entity":"incident:7e0d38d1-e5f5-4211-b2aa-3b142e4da80e",
		"text": "test with entity 1"
	}`)

	body := bytes.NewReader(payload)
	req := httptest.NewRequest("POST", "/comments", body)
	req.Header.Set("grpc-metadata-space", channelID)
	req.Header.Set("authorization", bearerToken)

	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)
	resp := w.Result()

	assert.Equal(t, http.StatusCreated, resp.StatusCode, "Status code")
	expectedLocation := "http://service.url/comments/38316161-3035-4864-ad30-6231392d3433"
	assert.Equal(t, expectedLocation, resp.Header.Get("Location"), "Location header")
}

func TestGetCommentDBMock(t *testing.T) {
	logger, _ := testutils.NewTestLogger()
	defer func() { _ = logger.Sync() }()

	bearerToken := "some valid Bearer token"
	channelID := "e27ddcd0-0e1f-4bc5-93df-f6f04155beec"
	orgID := "a897a407-e41b-4b14-924a-39f5d5a8038f"
	assetType := "comment"

	validator := new(mocks.ValidatorMock)
	validator.On("Validate", mock.AnythingOfType("comment.Comment")).Return(nil)

	events := new(mocks.EventServiceMock)
	queue := new(mocks.QueueMock)
	events.On("NewQueue", event.UUID(channelID), event.UUID(orgID)).Return(queue, nil)
	queue.On("AddCreateEvent", mock.AnythingOfType("comment.Comment"), assetType).Return(nil)
	queue.On("PublishEvents").Return(nil)

	couchMock, s := testutils.NewCouchDBMock(logger, validator, events)

	db := couchMock.NewDB()
	couchMock.ExpectDB().WithName(testutils.DatabaseName(channelID, assetType)).WillReturn(db)
	db.ExpectPut()

	db = couchMock.NewDB()
	couchMock.ExpectDB().WithName(testutils.DatabaseName(channelID, assetType)).WillReturn(db)

	dbDoc := comment.Comment{
		UUID:   "38316161-3035-4864-ad30-6231392d3433",
		Text:   "Test comment 1",
		Entity: entity.NewEntity("incident", "f49d5fd5-8da4-4779-b5ba-32e78aa2c444"),
		CreatedBy: &comment.UserInfo{
			UUID:           "8540d943-8ccd-4ff1-8a08-0c3aa338c58e",
			Name:           "Bob",
			Surname:        "Martin",
			OrgName:        "a897a407-e41b-4b14-924a-39f5d5a8038f.kompitech.com",
			OrgDisplayName: "Kompitech",
		},
		CreatedAt: "2021-04-01T12:34:56+02:00",
	}

	doc, err := kivikmock.Document(dbDoc)
	require.NoError(t, err)

	db.ExpectGet().WithDocID("38316161-3035-4864-ad30-6231392d3433").WillReturn(doc)

	c1 := comment.Comment{
		Text: "Test comment 1",
		CreatedBy: &comment.UserInfo{
			UUID:           "7e0d38d1-e5f5-4211-b2aa-3b142e4da80e",
			Name:           "Andy",
			Surname:        "Orange",
			OrgName:        "a897a407-e41b-4b14-924a-39f5d5a8038f.kompitech.com",
			OrgDisplayName: "Kompitech",
		},
	}

	as := new(mocks.AuthServiceMock)
	as.On("Enforce", assetType, auth.ReadAction, channelID, bearerToken).
		Return(true, nil)

	uuid, err := s.AddComment(c1, channelID, assetType)
	require.NoError(t, err)

	lister := listing.NewService(s)

	server := rest.NewServer(rest.Config{
		Addr:           "service.url",
		Logger:         logger,
		AuthService:    as,
		ListingService: lister,
	})

	req := httptest.NewRequest("GET", "/comments/"+uuid, nil)
	req.Header.Set("grpc-metadata-space", channelID)
	req.Header.Set("authorization", bearerToken)

	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)
	resp := w.Result()

	defer func() { _ = resp.Body.Close() }()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("could not read response: %v", err)
	}

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Status code")
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"), "Content-Type header")

	expectedJSON := `{
		"uuid":"38316161-3035-4864-ad30-6231392d3433",
		"text":"Test comment 1",
		"entity":"incident:f49d5fd5-8da4-4779-b5ba-32e78aa2c444",
		"created_by":{
			"uuid":"8540d943-8ccd-4ff1-8a08-0c3aa338c58e",
			"name":"Bob",
			"surname":"Martin",
			"org_name":"a897a407-e41b-4b14-924a-39f5d5a8038f.kompitech.com",
			"org_display_name":"Kompitech"
		},
		"created_at":"2021-04-01T12:34:56+02:00"
	}`
	require.JSONEq(t, expectedJSON, string(b), "response does not match")
}

func TestAddCommentMemoryStorage(t *testing.T) {
	logger, _ := testutils.NewTestLogger()
	defer func() { _ = logger.Sync() }()

	channelID := "e27ddcd0-0e1f-4bc5-93df-f6f04155beec"
	bearerToken := "some valid Bearer token"

	rand := strings.NewReader("81aa058d-0b19-43e9-82ae-a7bca2457f10") // pseudo-random seed
	s := &memory.Storage{
		Clock: testutils.FixedClock{},
		Rand:  rand,
	}
	adder := adding.NewService(s)

	as := new(mocks.AuthServiceMock)
	as.On("Enforce", "comment", auth.CreateAction, channelID, bearerToken).
		Return(true, nil)

	us := new(mocks.UserServiceMock)
	mockUserData := user.BasicInfo{
		UUID: "2af4f493-0bd5-4513-b440-6cbb465feadb",
		Name: "Some test user 1",
	}
	us.On("UserBasicInfo", mock.AnythingOfType("*http.Request")).
		Return(mockUserData, nil)

	pv, err := validation.NewPayloadValidator()
	require.NoError(t, err)

	server := rest.NewServer(rest.Config{
		Addr:             "service.url",
		AuthService:      as,
		UserService:      us,
		Logger:           logger,
		AddingService:    adder,
		PayloadValidator: pv,
	})

	payload := []byte(`{"entity":"incident:7e0d38d1-e5f5-4211-b2aa-3b142e4da80e", "text": "test with entity 1"}`)

	body := bytes.NewReader(payload)
	req := httptest.NewRequest("POST", "/comments", body)
	req.Header.Set("grpc-metadata-space", channelID)
	req.Header.Set("authorization", bearerToken)

	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)
	resp := w.Result()

	assert.Equal(t, http.StatusCreated, resp.StatusCode, "Status code")
	expectedLocation := "http://service.url/comments/38316161-3035-4864-ad30-6231392d3433"
	assert.Equal(t, expectedLocation, resp.Header.Get("Location"), "Location header")
}

func TestGetCommentMemoryStorage(t *testing.T) {
	logger, _ := testutils.NewTestLogger()
	defer func() { _ = logger.Sync() }()

	rand := strings.NewReader("81aa058d-0b19-43e9-82ae-a7bca2457f10") // pseudo-random seed
	s := &memory.Storage{
		Clock: testutils.FixedClock{},
		Rand:  rand,
	}

	c1 := comment.Comment{
		Text:   "Test comment 1",
		Entity: entity.NewEntity("incident", "f49d5fd5-8da4-4779-b5ba-32e78aa2c444"),
	}

	bearerToken := "some valid Bearer token"
	channelID := "e27ddcd0-0e1f-4bc5-93df-f6f04155beec"
	assetType := "comment"

	as := new(mocks.AuthServiceMock)
	as.On("Enforce", assetType, auth.ReadAction, channelID, bearerToken).
		Return(true, nil)

	uuid, err := s.AddComment(c1, channelID, assetType)
	require.NoError(t, err)

	lister := listing.NewService(s)

	server := rest.NewServer(rest.Config{
		Addr:           "service.url",
		Logger:         logger,
		AuthService:    as,
		ListingService: lister,
	})

	req := httptest.NewRequest("GET", "/comments/"+uuid, nil)
	req.Header.Set("grpc-metadata-space", channelID)
	req.Header.Set("authorization", bearerToken)

	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)
	resp := w.Result()

	defer func() { _ = resp.Body.Close() }()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("could not read response: %v", err)
	}

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Status code")
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"), "Content-Type header")

	expectedJSON := `{
		"uuid":"38316161-3035-4864-ad30-6231392d3433",
		"text":"Test comment 1",
		"entity":"incident:f49d5fd5-8da4-4779-b5ba-32e78aa2c444",
		"created_at":"2021-04-01T12:34:56+02:00"
	}`
	require.JSONEq(t, expectedJSON, string(b), "response does not match")
}
