package rest

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KompiTech/itsm-commenting-service/pkg/domain/user"
	"github.com/KompiTech/itsm-commenting-service/pkg/http/rest/validation"
	"github.com/KompiTech/itsm-commenting-service/pkg/mocks"
	"github.com/KompiTech/itsm-commenting-service/pkg/repository/couchdb"
	pvalidation "github.com/KompiTech/itsm-commenting-service/pkg/validation"
	"github.com/KompiTech/itsm-commenting-service/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAddCommentHandler(t *testing.T) {
	logger := testutils.NewTestLogger()
	defer func() { _ = logger.Sync() }()

	mockUserData := user.BasicInfo{
		UUID: "2af4f493-0bd5-4513-b440-6cbb465feadb",
		Name: "Some test user 1",
	}

	channelID := "e27ddcd0-0e1f-4bc5-93df-f6f04155beec"

	t.Run("when channelID is not set (ie. grpc-metadata-space header is missing)", func(t *testing.T) {
		us := new(mocks.UserServiceMock)
		us.On("UserBasicInfo", mock.AnythingOfType("*http.Request")).
			Return(mockUserData, nil)

		pv, err := validation.NewPayloadValidator()
		require.NoError(t, err)

		server := NewServer(Config{
			Addr:             "service.url",
			Logger:           logger,
			UserService:      us,
			PayloadValidator: pv,
		})

		payload := []byte(`{
			"entity":"incident:7e0d38d1-e5f5-4211-b2aa-3b142e4da80e",
			"text": "test with entity 1",
			"external_id": "someExternalID"
		}`)

		body := bytes.NewReader(payload)
		req := httptest.NewRequest("POST", "/comments", body)

		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)
		resp := w.Result()

		defer func() { _ = resp.Body.Close() }()
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("could not read response: %v", err)
		}

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "Status code")
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"), "Content-Type header")

		expectedJSON := `{"error":"'grpc-metadata-space' header missing or invalid"}`
		assert.JSONEq(t, expectedJSON, string(b), "response does not match")
	})

	t.Run("when request is not valid JSON", func(t *testing.T) {
		us := new(mocks.UserServiceMock)
		us.On("UserBasicInfo", mock.AnythingOfType("*http.Request")).
			Return(mockUserData, nil)

		pv, err := validation.NewPayloadValidator()
		require.NoError(t, err)

		server := NewServer(Config{
			Addr:             "service.url",
			Logger:           logger,
			UserService:      us,
			PayloadValidator: pv,
		})

		payload := []byte(`{"invalid json request"}`)

		body := bytes.NewReader(payload)
		req := httptest.NewRequest("POST", "/comments", body)

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

	t.Run("when request is not valid ('uuid' key present, empty 'text' key)", func(t *testing.T) {
		us := new(mocks.UserServiceMock)
		us.On("UserBasicInfo", mock.AnythingOfType("*http.Request")).
			Return(mockUserData, nil)

		pv, err := validation.NewPayloadValidator()
		require.NoError(t, err)

		server := NewServer(Config{
			Addr:             "service.url",
			Logger:           logger,
			UserService:      us,
			PayloadValidator: pv,
		})

		payload := []byte(`{
			"uuid": "1e88630d-2457-4f60-a66c-34a542a2e1f4",
			"entity":"incident:7e0d38d1-e5f5-4211-b2aa-3b142e4da80e",
			"text": ""
		}`)

		body := bytes.NewReader(payload)
		req := httptest.NewRequest("POST", "/comments", body)
		req.Header.Set("grpc-metadata-space", channelID)

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

		expectedJSON := `{"error":"/: additional properties are not allowed\n/text: regexp pattern \\S mismatch on string: "}`
		assert.JSONEq(t, expectedJSON, string(b), "response does not match")
	})

	t.Run("when validator fails (ie. returns general error", func(t *testing.T) {
		us := new(mocks.UserServiceMock)
		us.On("UserBasicInfo", mock.AnythingOfType("*http.Request")).
			Return(mockUserData, nil)

		pv := new(mocks.PayloadValidatorMock)
		pv.On("ValidatePayload", mock.AnythingOfType("[]uint8"), mock.AnythingOfType("string")).
			Return(pvalidation.NewErrGeneral(errors.New("could not open schema file")))

		server := NewServer(Config{
			Addr:             "service.url",
			Logger:           logger,
			UserService:      us,
			PayloadValidator: pv,
		})

		payload := []byte(`{
			"uuid": "1e88630d-2457-4f60-a66c-34a542a2e1f4",
			"entity":"incident:7e0d38d1-e5f5-4211-b2aa-3b142e4da80e",
			"text": "Comment 1"
		}`)

		body := bytes.NewReader(payload)
		req := httptest.NewRequest("POST", "/comments", body)
		req.Header.Set("grpc-metadata-space", channelID)

		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)
		resp := w.Result()

		defer func() { _ = resp.Body.Close() }()
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("could not read response: %v", err)
		}

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode, "Status code")
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"), "Content-Type header")

		expectedJSON := `{"error":"general error: could not open schema file"}`
		assert.JSONEq(t, expectedJSON, string(b), "response does not match")
	})

	t.Run("when comment was not stored yet", func(t *testing.T) {
		us := new(mocks.UserServiceMock)
		us.On("UserBasicInfo", mock.AnythingOfType("*http.Request")).
			Return(mockUserData, nil)

		adder := new(mocks.AddingMock)
		assetType := "comment"
		adder.On("AddComment", mock.AnythingOfType("comment.Comment"), channelID, assetType).
			Return("38316161-3035-4864-ad30-6231392d3433", nil)

		pv, err := validation.NewPayloadValidator()
		require.NoError(t, err)

		server := NewServer(Config{
			Addr:             "service.url",
			Logger:           logger,
			UserService:      us,
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

		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)
		resp := w.Result()

		assert.Equal(t, http.StatusCreated, resp.StatusCode, "Status code")
		expectedLocation := "http://service.url/comments/38316161-3035-4864-ad30-6231392d3433"
		assert.Equal(t, expectedLocation, resp.Header.Get("Location"), "Location header")
	})

	t.Run("when repository returns conflict error (ie. trying to add already stored comment)", func(t *testing.T) {
		us := new(mocks.UserServiceMock)
		us.On("UserBasicInfo", mock.AnythingOfType("*http.Request")).
			Return(mockUserData, nil)

		adder := new(mocks.AddingMock)
		assetType := "comment"
		adder.On("AddComment", mock.AnythingOfType("comment.Comment"), channelID, assetType).
			Return("", couchdb.ErrorConflict("Comment already exists"))

		pv, err := validation.NewPayloadValidator()
		require.NoError(t, err)

		server := NewServer(Config{
			Addr:             "service.url",
			Logger:           logger,
			UserService:      us,
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

		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)
		resp := w.Result()

		defer func() { _ = resp.Body.Close() }()
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("could not read response: %v", err)
		}

		assert.Equal(t, http.StatusConflict, resp.StatusCode, "Status code")
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"), "Content-Type header")

		expectedJSON := `{"error":"Comment already exists"}`
		assert.JSONEq(t, expectedJSON, string(b), "response does not match")
	})

	t.Run("when repository returns some other general error", func(t *testing.T) {
		us := new(mocks.UserServiceMock)
		us.On("UserBasicInfo", mock.AnythingOfType("*http.Request")).
			Return(mockUserData, nil)

		adder := new(mocks.AddingMock)
		assetType := "comment"
		adder.On("AddComment", mock.AnythingOfType("comment.Comment"), channelID, assetType).
			Return("", errors.New("some error occurred"))

		pv, err := validation.NewPayloadValidator()
		require.NoError(t, err)

		server := NewServer(Config{
			Addr:             "service.url",
			Logger:           logger,
			UserService:      us,
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

		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)
		resp := w.Result()

		defer func() { _ = resp.Body.Close() }()
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("could not read response: %v", err)
		}

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode, "Status code")
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"), "Content-Type header")

		expectedJSON := `{"error":"some error occurred"}`
		assert.JSONEq(t, expectedJSON, string(b), "response does not match")
	})

	t.Run("when user service failed to retrieve user info and put it in the request context", func(t *testing.T) {
		us := new(mocks.UserServiceMock)
		us.On("UserBasicInfo", mock.AnythingOfType("*http.Request")).
			Return(user.BasicInfo{}, errors.New("some user service error"))

		pv, err := validation.NewPayloadValidator()
		require.NoError(t, err)

		server := NewServer(Config{
			Addr:             "service.url",
			Logger:           logger,
			UserService:      us,
			PayloadValidator: pv,
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

		defer func() { _ = resp.Body.Close() }()
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("could not read response: %v", err)
		}

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode, "Status code")
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"), "Content-Type header")

		expectedJSON := `{"error":"could not retrieve correct user info from user service: some user service error"}`
		assert.JSONEq(t, expectedJSON, string(b), "response does not match")
	})

	// worknote
	t.Run("when worknote was not stored yet", func(t *testing.T) {
		us := new(mocks.UserServiceMock)
		us.On("UserBasicInfo", mock.AnythingOfType("*http.Request")).
			Return(mockUserData, nil)

		adder := new(mocks.AddingMock)
		assetType := "worknote"
		adder.On("AddComment", mock.AnythingOfType("comment.Comment"), channelID, assetType).
			Return("38316161-3035-4864-ad30-6231392d3433", nil)

		pv, err := validation.NewPayloadValidator()
		require.NoError(t, err)

		server := NewServer(Config{
			Addr:             "service.url",
			Logger:           logger,
			UserService:      us,
			AddingService:    adder,
			PayloadValidator: pv,
		})

		payload := []byte(`{
			"entity":"incident:7e0d38d1-e5f5-4211-b2aa-3b142e4da80e",
			"text": "test with worknote for entity 1"
		}`)

		body := bytes.NewReader(payload)
		req := httptest.NewRequest("POST", "/worknotes", body)
		req.Header.Set("grpc-metadata-space", channelID)

		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)
		resp := w.Result()

		assert.Equal(t, http.StatusCreated, resp.StatusCode, "Status code")
		expectedLocation := "http://service.url/worknotes/38316161-3035-4864-ad30-6231392d3433"
		assert.Equal(t, expectedLocation, resp.Header.Get("Location"), "Location header")
	})

}
