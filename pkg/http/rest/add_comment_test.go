package rest

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KompiTech/itsm-commenting-service/pkg/mocks"
	"github.com/KompiTech/itsm-commenting-service/pkg/repository/couchdb"
	"github.com/KompiTech/itsm-commenting-service/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAddCommentHandler(t *testing.T) {
	logger := testutils.NewTestLogger()
	defer func() { _ = logger.Sync() }()

	t.Run("when comment exists", func(t *testing.T) {
		adder := new(mocks.AddingMock)
		adder.On("AddComment", mock.AnythingOfType("comment.Comment")).
			Return("38316161-3035-4864-ad30-6231392d3433", nil)

		server := NewServer(Config{
			Addr:          "service.url",
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
	})

	t.Run("when repository returns conflict error (ie. trying to add already stored comment)", func(t *testing.T) {
		adder := new(mocks.AddingMock)
		adder.On("AddComment", mock.AnythingOfType("comment.Comment")).
			Return("", couchdb.ErrorConflict("Comment already exists"))

		server := NewServer(Config{
			Addr:          "service.url",
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
		adder := new(mocks.AddingMock)
		adder.On("AddComment", mock.AnythingOfType("comment.Comment")).
			Return("", errors.New("some error occurred"))

		server := NewServer(Config{
			Addr:          "service.url",
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
}
