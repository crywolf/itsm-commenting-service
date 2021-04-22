package rest

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment"
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/entity"
	"github.com/KompiTech/itsm-commenting-service/pkg/mocks"
	"github.com/KompiTech/itsm-commenting-service/pkg/repository/couchdb"
	"github.com/KompiTech/itsm-commenting-service/testutils"
	"github.com/stretchr/testify/assert"
)

func TestGetCommentHandler(t *testing.T) {
	logger := testutils.NewTestLogger()
	defer func() { _ = logger.Sync() }()

	t.Run("when comment exists", func(t *testing.T) {
		uuid := "cb2fe2a7-ab9f-4f6d-9fd6-c7c209403cf0"
		retC := comment.Comment{
			Text:   "Test comment 1",
			Entity: entity.NewEntity("incident", "f49d5fd5-8da4-4779-b5ba-32e78aa2c444"),
			UUID:   uuid,
			CreatedBy: &comment.CreatedBy{
				UUID:    "8540d943-8ccd-4ff1-8a08-0c3aa338c58e",
				Name:    "Alice",
				Surname: "Cooper",
			},
			CreatedAt: "2021-04-01T12:34:56+02:00",
		}

		lister := new(mocks.ListingMock)
		lister.On("GetComment", uuid).
			Return(retC, nil)

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
			"uuid":"cb2fe2a7-ab9f-4f6d-9fd6-c7c209403cf0",
			"text":"Test comment 1",
			"entity":"incident:f49d5fd5-8da4-4779-b5ba-32e78aa2c444",
			"created_by":{
				"uuid":"8540d943-8ccd-4ff1-8a08-0c3aa338c58e",
				"name":"Alice",
				"surname":"Cooper"
			},
			"created_at":"2021-04-01T12:34:56+02:00"
		}`
		assert.JSONEq(t, expectedJSON, string(b), "response does not match")
	})

	t.Run("when comment does not exist", func(t *testing.T) {
		uuid := "someNonexistentUUID"

		lister := new(mocks.ListingMock)
		lister.On("GetComment", uuid).
			Return(comment.Comment{}, couchdb.ErrorNorFound("comment not found"))

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

		assert.Equal(t, http.StatusNotFound, resp.StatusCode, "Status code")
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"), "Content-Type header")

		expectedJSON := `{"error":"comment not found"}`
		assert.JSONEq(t, expectedJSON, string(b), "response does not match")
	})

	t.Run("when repository returns some other error", func(t *testing.T) {
		uuid := "someNonexistentUUID"

		lister := new(mocks.ListingMock)
		lister.On("GetComment", uuid).
			Return(comment.Comment{}, errors.New("some error occurred"))

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

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode, "Status code")
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"), "Content-Type header")

		expectedJSON := `{"error":"some error occurred"}`
		assert.JSONEq(t, expectedJSON, string(b), "response does not match")
	})
}
