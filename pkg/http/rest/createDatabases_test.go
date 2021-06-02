package rest

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KompiTech/itsm-commenting-service/pkg/http/rest/auth"
	"github.com/KompiTech/itsm-commenting-service/pkg/http/rest/validation"
	"github.com/KompiTech/itsm-commenting-service/pkg/mocks"
	"github.com/KompiTech/itsm-commenting-service/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateDatabasesHandler(t *testing.T) {
	logger, _ := testutils.NewTestLogger()
	defer func() { _ = logger.Sync() }()

	channelID := "e27ddcd0-0e1f-4bc5-93df-f6f04155beec"
	bearerToken := "some valid Bearer token"

	t.Run("when request is not valid JSON", func(t *testing.T) {
		as := new(mocks.AuthServiceMock)
		as.On("Enforce", "database", auth.UpdateAction, bearerToken).
			Return(true, nil)

		pv, err := validation.NewPayloadValidator()
		require.NoError(t, err)

		server := NewServer(Config{
			Addr:             "service.url",
			Logger:           logger,
			AuthService:      as,
			PayloadValidator: pv,
		})

		payload := []byte(`{"invalid json request"}`)

		body := bytes.NewReader(payload)
		req := httptest.NewRequest("POST", "/databases", body)
		req.Header.Set("authorization", bearerToken)

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

		expectedJSON := `{"error":"error parsing JSON bytes: invalid character '}' after object key"}`
		assert.JSONEq(t, expectedJSON, string(b), "response does not match")
	})

	t.Run("when request is not valid ('channel_id' missing)", func(t *testing.T) {
		as := new(mocks.AuthServiceMock)
		as.On("Enforce", "database", auth.UpdateAction, bearerToken).
			Return(true, nil)

		pv, err := validation.NewPayloadValidator()
		require.NoError(t, err)

		server := NewServer(Config{
			Addr:             "service.url",
			AuthService:      as,
			Logger:           logger,
			PayloadValidator: pv,
		})

		payload := []byte(`{}`)

		body := bytes.NewReader(payload)
		req := httptest.NewRequest("POST", "/databases", body)
		req.Header.Set("authorization", bearerToken)

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

		expectedJSON := `{"error":"/: 'channel_id' value is required"}`
		assert.JSONEq(t, expectedJSON, string(b), "response does not match")
	})

	t.Run("when databases already exist", func(t *testing.T) {
		couchMock, s := testutils.NewCouchDBMock(logger, nil, nil)
		couchMock.ExpectDBExists().WithName(testutils.DatabaseName(channelID, "comment")).WillReturn(true)
		couchMock.ExpectDBExists().WithName(testutils.DatabaseName(channelID, "worknote")).WillReturn(true)

		as := new(mocks.AuthServiceMock)
		as.On("Enforce", "database", auth.UpdateAction, bearerToken).
			Return(true, nil)

		pv, err := validation.NewPayloadValidator()
		require.NoError(t, err)

		server := NewServer(Config{
			Addr:              "service.url",
			Logger:            logger,
			AuthService:       as,
			RepositoryService: s,
			PayloadValidator:  pv,
		})

		payload := []byte(`{"channel_id":"e27ddcd0-0e1f-4bc5-93df-f6f04155beec"}`)

		body := bytes.NewReader(payload)
		req := httptest.NewRequest("POST", "/databases", body)
		req.Header.Set("authorization", bearerToken)

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
		couchMock, s := testutils.NewCouchDBMock(logger, nil, nil)

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

		as := new(mocks.AuthServiceMock)
		as.On("Enforce", "database", auth.UpdateAction, bearerToken).
			Return(true, nil)

		pv, err := validation.NewPayloadValidator()
		require.NoError(t, err)

		server := NewServer(Config{
			Addr:              "service.url",
			Logger:            logger,
			AuthService:       as,
			RepositoryService: s,
			PayloadValidator:  pv,
		})

		payload := []byte(`{"channel_id":"e27ddcd0-0e1f-4bc5-93df-f6f04155beec"}`)

		body := bytes.NewReader(payload)
		req := httptest.NewRequest("POST", "/databases", body)
		req.Header.Set("authorization", bearerToken)

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
