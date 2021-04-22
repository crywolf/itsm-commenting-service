package rest

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
	logger := testutils.NewTestLogger()
	defer func() { _ = logger.Sync() }()

	couchMock, s := testutils.NewCouchDBMock(logger)

	db := couchMock.NewDB()
	couchMock.ExpectDB().WithName("comments").WillReturn(db)
	db.ExpectPut()

	us := new(mocks.UserServiceMock)
	mockUserData := user.BasicInfo{
		UUID:    "2af4f493-0bd5-4513-b440-6cbb465feadb",
		Name:    "Alice",
		Surname: "Cooper",
	}
	us.On("UserBasicInfo", mock.AnythingOfType("*http.Request")).
		Return(mockUserData, nil)

	adder := adding.NewService(s)

	server := NewServer(Config{
		Addr:          "service.url",
		UserService:   us,
		Logger:        logger,
		AddingService: adder,
	})

	payload := []byte(`{
		"entity":"incident:7e0d38d1-e5f5-4211-b2aa-3b142e4da80e",
		"text": "test with entity 1"
	}`)

	body := bytes.NewReader(payload)
	req := httptest.NewRequest("POST", "/comments", body)

	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)
	resp := w.Result()

	assert.Equal(t, http.StatusCreated, resp.StatusCode, "Status code")
	expectedLocation := "http://service.url/comments/38316161-3035-4864-ad30-6231392d3433"
	assert.Equal(t, expectedLocation, resp.Header.Get("Location"), "Location header")
}

func TestGetCommentDBMock(t *testing.T) {
	logger := testutils.NewTestLogger()
	defer func() { _ = logger.Sync() }()

	couchMock, s := testutils.NewCouchDBMock(logger)

	db := couchMock.NewDB()
	couchMock.ExpectDB().WithName("comments").WillReturn(db)
	db.ExpectPut()

	db = couchMock.NewDB()
	couchMock.ExpectDB().WithName("comments").WillReturn(db)

	dbDoc := comment.Comment{
		UUID:   "38316161-3035-4864-ad30-6231392d3433",
		Text:   "Test comment 1",
		Entity: entity.NewEntity("incident", "f49d5fd5-8da4-4779-b5ba-32e78aa2c444"),
		CreatedBy: &comment.CreatedBy{
			UUID:    "8540d943-8ccd-4ff1-8a08-0c3aa338c58e",
			Name:    "Bob",
			Surname: "Martin",
		},
		CreatedAt: "2021-04-01T12:34:56+02:00",
	}

	doc, err := kivikmock.Document(dbDoc)
	require.NoError(t, err)

	db.ExpectGet().WithDocID("38316161-3035-4864-ad30-6231392d3433").WillReturn(doc)

	c1 := comment.Comment{
		//Text:   "Test comment 1",
		//Entity: entity.NewEntity("incident", "f49d5fd5-8da4-4779-b5ba-32e78aa2c444"),
	}

	uuid, err := s.AddComment(c1)
	require.NoError(t, err)

	lister := listing.NewService(s)

	server := NewServer(Config{
		Addr:           "service.url",
		Logger:         logger,
		ListingService: lister,
	})

	req := httptest.NewRequest("GET", "/comments/"+uuid, nil)

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
			"surname": "Martin"
		},
		"created_at":"2021-04-01T12:34:56+02:00"
	}`
	require.JSONEq(t, expectedJSON, string(b), "response does not match")
}

//////////////////////////////

type AdderStub struct{}

func (a AdderStub) AddComment(_ comment.Comment) (id string, err error) {
	id = "38316161-3035-4864-ad30-6231392d3433"
	return id, err
}

func TestAddCommentAdderStub(t *testing.T) {
	logger := testutils.NewTestLogger()
	defer func() { _ = logger.Sync() }()

	adder := &AdderStub{}

	us := new(mocks.UserServiceMock)
	mockUserData := user.BasicInfo{
		UUID: "2af4f493-0bd5-4513-b440-6cbb465feadb",
		Name: "Some test user 1",
	}
	us.On("UserBasicInfo", mock.AnythingOfType("*http.Request")).
		Return(mockUserData, nil)

	server := NewServer(Config{
		Addr:          "service.url",
		UserService:   us,
		Logger:        logger,
		AddingService: adder,
	})

	payload := []byte(`{
		"entity":"incident:7e0d38d1-e5f5-4211-b2aa-3b142e4da80e",
		"text": "test with entity 1"
	}`)

	body := bytes.NewReader(payload)
	req := httptest.NewRequest("POST", "/comments", body)

	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)
	resp := w.Result()

	assert.Equal(t, http.StatusCreated, resp.StatusCode, "Status code")
	expectedLocation := "http://service.url/comments/38316161-3035-4864-ad30-6231392d3433"
	assert.Equal(t, expectedLocation, resp.Header.Get("Location"), "Location header")
}

/////////////////////////

func TestAddCommentMemoryStorage(t *testing.T) {
	logger := testutils.NewTestLogger()
	defer func() { _ = logger.Sync() }()

	rand := strings.NewReader("81aa058d-0b19-43e9-82ae-a7bca2457f10") // pseudo-random seed
	s := &memory.Storage{
		Clock: testutils.FixedClock{},
		Rand:  rand,
	}
	adder := adding.NewService(s)

	us := new(mocks.UserServiceMock)
	mockUserData := user.BasicInfo{
		UUID: "2af4f493-0bd5-4513-b440-6cbb465feadb",
		Name: "Some test user 1",
	}
	us.On("UserBasicInfo", mock.AnythingOfType("*http.Request")).
		Return(mockUserData, nil)

	server := NewServer(Config{
		Addr:          "service.url",
		UserService:   us,
		Logger:        logger,
		AddingService: adder,
	})

	payload := []byte(`{"entity":"incident:7e0d38d1-e5f5-4211-b2aa-3b142e4da80e", "text": "test with entity 1"}`)

	body := bytes.NewReader(payload)
	req := httptest.NewRequest("POST", "/comments", body)

	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)
	resp := w.Result()

	assert.Equal(t, http.StatusCreated, resp.StatusCode, "Status code")
	expectedLocation := "http://service.url/comments/38316161-3035-4864-ad30-6231392d3433"
	assert.Equal(t, expectedLocation, resp.Header.Get("Location"), "Location header")
}

func TestGetCommentMemoryStorage(t *testing.T) {
	logger := testutils.NewTestLogger()
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

	uuid, err := s.AddComment(c1)
	require.NoError(t, err)

	lister := listing.NewService(s)

	server := NewServer(Config{
		Addr:           "service.url",
		Logger:         logger,
		ListingService: lister,
	})

	req := httptest.NewRequest("GET", "/comments/"+uuid, nil)

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
