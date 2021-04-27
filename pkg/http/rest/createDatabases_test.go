package rest

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KompiTech/itsm-commenting-service/testutils"
	"github.com/stretchr/testify/assert"
)

func TestCreateDatabasesHandler(t *testing.T) {
	logger := testutils.NewTestLogger()
	defer func() { _ = logger.Sync() }()

	channelID := "e27ddcd0-0e1f-4bc5-93df-f6f04155beec"

	t.Run("when request is not valid", func(t *testing.T) {
		server := NewServer(Config{
			Addr:   "service.url",
			Logger: logger,
		})

		payload := []byte(`{"invalid json request"}`)

		body := bytes.NewReader(payload)
		req := httptest.NewRequest("POST", "/databases", body)

		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)
		resp := w.Result()

		defer func() { _ = resp.Body.Close() }()
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("could not read response: %v", err)
		}

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "Status code")
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"), "Content-Type header")

		expectedJSON := `{"error":"could not decode JSON from request: invalid character '}' after object key"}`
		assert.JSONEq(t, expectedJSON, string(b), "response does not match")
	})

	t.Run("when databases already exist", func(t *testing.T) {
		couchMock, s := testutils.NewCouchDBMock(logger)

		couchMock.ExpectDBExists().WithName(testutils.DatabaseName(channelID, "comment")).WillReturn(true)
		couchMock.ExpectDBExists().WithName(testutils.DatabaseName(channelID, "worknote")).WillReturn(true)

		server := NewServer(Config{
			Addr:              "service.url",
			Logger:            logger,
			RepositoryService: s,
		})

		payload := []byte(`{"channel_id":"e27ddcd0-0e1f-4bc5-93df-f6f04155beec"}`)

		body := bytes.NewReader(payload)
		req := httptest.NewRequest("POST", "/databases", body)

		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)
		resp := w.Result()

		defer func() { _ = resp.Body.Close() }()
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("could not read response: %v", err)
		}

		assert.Equal(t, http.StatusNoContent, resp.StatusCode, "Status code")
		assert.Empty(t, b)
	})

	t.Run("when databases do not exist", func(t *testing.T) {
		couchMock, s := testutils.NewCouchDBMock(logger)

		// comments
		couchMock.ExpectDBExists().WithName(testutils.DatabaseName(channelID, "comment")).WillReturn(false)
		couchMock.ExpectCreateDB().WithName(testutils.DatabaseName(channelID, "comment"))
		db := couchMock.NewDB()
		couchMock.ExpectDB().WithName(testutils.DatabaseName(channelID, "comment")).WillReturn(db)
		db.ExpectCreateIndex()

		// worknotes
		couchMock.ExpectDBExists().WithName(testutils.DatabaseName(channelID, "worknote")).WillReturn(false)
		couchMock.ExpectCreateDB().WithName(testutils.DatabaseName(channelID, "worknote"))
		couchMock.ExpectDB().WithName(testutils.DatabaseName(channelID, "worknote")).WillReturn(db)
		db.ExpectCreateIndex()

		server := NewServer(Config{
			Addr:              "service.url",
			Logger:            logger,
			RepositoryService: s,
		})

		payload := []byte(`{"channel_id":"e27ddcd0-0e1f-4bc5-93df-f6f04155beec"}`)

		body := bytes.NewReader(payload)
		req := httptest.NewRequest("POST", "/databases", body)

		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)
		resp := w.Result()

		defer func() { _ = resp.Body.Close() }()
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("could not read response: %v", err)
		}

		assert.Equal(t, http.StatusCreated, resp.StatusCode, "Status code")
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"), "Content-Type header")

		expectedJSON := `{"message":"databases were successfully created"}`
		assert.JSONEq(t, expectedJSON, string(b), "response does not match")
	})
}
