package rest

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment"
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/entity"
	"github.com/KompiTech/itsm-commenting-service/pkg/http/rest/auth"
	"github.com/KompiTech/itsm-commenting-service/pkg/mocks"
	"github.com/KompiTech/itsm-commenting-service/pkg/repository/couchdb"
	"github.com/KompiTech/itsm-commenting-service/testutils"
	"github.com/stretchr/testify/assert"
)

func TestGetCommentHandler(t *testing.T) {
	logger, _ := testutils.NewTestLogger()
	defer func() { _ = logger.Sync() }()

	channelID := "e27ddcd0-0e1f-4bc5-93df-f6f04155beec"
	bearerToken := "some valid Bearer token"

	t.Run("when 'authorization' header with Bearer token is missing", func(t *testing.T) {
		server := NewServer(Config{
			Addr:   "service.url",
			Logger: logger,
		})

		uuid := "cb2fe2a7-ab9f-4f6d-9fd6-c7c209403cf0"
		req := httptest.NewRequest("GET", "/comments/"+uuid, nil)

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

		expectedJSON := `{"error":"'authorization' header missing or invalid"}`
		assert.JSONEq(t, expectedJSON, string(b), "response does not match")
	})

	t.Run("when authorization service returns error", func(t *testing.T) {
		as := new(mocks.AuthServiceMock)
		assetType := comment.AssetTypeComment
		as.On("Enforce", assetType.String(), auth.ReadAction, channelID, bearerToken).
			Return(false, errors.New("some authorization service error"))

		server := NewServer(Config{
			Addr:        "service.url",
			Logger:      logger,
			AuthService: as,
		})

		uuid := "cb2fe2a7-ab9f-4f6d-9fd6-c7c209403cf0"
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

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode, "Status code")
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"), "Content-Type header")

		expectedJSON := `{"error":"Authorization failed: some authorization service error"}`
		assert.JSONEq(t, expectedJSON, string(b), "response does not match")
	})

	t.Run("when channelID is not set (ie. grpc-metadata-space header is missing)", func(t *testing.T) {
		as := new(mocks.AuthServiceMock)

		server := NewServer(Config{
			Addr:        "service.url",
			Logger:      logger,
			AuthService: as,
		})

		uuid := "cb2fe2a7-ab9f-4f6d-9fd6-c7c209403cf0"
		req := httptest.NewRequest("GET", "/comments/"+uuid, nil)
		req.Header.Set("authorization", bearerToken)

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

	t.Run("when comment exists", func(t *testing.T) {
		uuid := "cb2fe2a7-ab9f-4f6d-9fd6-c7c209403cf0"
		retC := comment.Comment{
			Text:   "Test comment 1",
			Entity: entity.NewEntity("incident", "f49d5fd5-8da4-4779-b5ba-32e78aa2c444"),
			UUID:   uuid,
			CreatedBy: &comment.UserInfo{
				UUID:           "8540d943-8ccd-4ff1-8a08-0c3aa338c58e",
				Name:           "Alice",
				Surname:        "Cooper",
				OrgName:        "a897a407-e41b-4b14-924a-39f5d5a8038f.kompitech.com",
				OrgDisplayName: "Kompitech",
			},
			CreatedAt: "2021-04-01T12:34:56+02:00",
		}

		assetType := comment.AssetTypeComment
		as := new(mocks.AuthServiceMock)
		as.On("Enforce", assetType.String(), auth.ReadAction, channelID, bearerToken).
			Return(true, nil)

		lister := new(mocks.ListingMock)
		lister.On("GetComment", uuid, channelID, assetType).
			Return(retC, nil)

		server := NewServer(Config{
			Addr:                    "service.url",
			Logger:                  logger,
			AuthService:             as,
			ListingService:          lister,
			ExternalLocationAddress: "http://service.url",
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
			"uuid":"cb2fe2a7-ab9f-4f6d-9fd6-c7c209403cf0",
			"text":"Test comment 1",
			"entity":"incident:f49d5fd5-8da4-4779-b5ba-32e78aa2c444",
			"created_by":{
				"uuid":"8540d943-8ccd-4ff1-8a08-0c3aa338c58e",
				"name":"Alice",
				"surname":"Cooper",
				"org_name":"a897a407-e41b-4b14-924a-39f5d5a8038f.kompitech.com",
				"org_display_name":"Kompitech"
			},
			"created_at":"2021-04-01T12:34:56+02:00",
			"_links":{
				"self":{"href":"http://service.url/comments/cb2fe2a7-ab9f-4f6d-9fd6-c7c209403cf0"},
				"MarkCommentAsReadByUser":{"href":"http://service.url/comments/cb2fe2a7-ab9f-4f6d-9fd6-c7c209403cf0/read_by"}
			}
		}`
		assert.JSONEq(t, expectedJSON, string(b), "response does not match")
	})

	t.Run("when comment does not exist", func(t *testing.T) {
		uuid := "someNonexistentUUID"
		assetType := comment.AssetTypeComment

		as := new(mocks.AuthServiceMock)
		as.On("Enforce", assetType.String(), auth.ReadAction, channelID, bearerToken).
			Return(true, nil)

		lister := new(mocks.ListingMock)
		lister.On("GetComment", uuid, channelID, assetType).
			Return(comment.Comment{}, couchdb.ErrorNorFound("comment not found"))

		server := NewServer(Config{
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

		assert.Equal(t, http.StatusNotFound, resp.StatusCode, "Status code")
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"), "Content-Type header")

		expectedJSON := `{"error":"comment not found"}`
		assert.JSONEq(t, expectedJSON, string(b), "response does not match")
	})

	t.Run("when repository returns some other error", func(t *testing.T) {
		uuid := "someNonexistentUUID"
		assetType := comment.AssetTypeComment

		as := new(mocks.AuthServiceMock)
		as.On("Enforce", assetType.String(), auth.ReadAction, channelID, bearerToken).
			Return(true, nil)

		lister := new(mocks.ListingMock)
		lister.On("GetComment", uuid, channelID, assetType).
			Return(comment.Comment{}, errors.New("some error occurred"))

		server := NewServer(Config{
			Addr:           "service.url",
			AuthService:    as,
			Logger:         logger,
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

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode, "Status code")
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"), "Content-Type header")

		expectedJSON := `{"error":"some error occurred"}`
		assert.JSONEq(t, expectedJSON, string(b), "response does not match")
	})

	// worknote
	t.Run("when worknote exists", func(t *testing.T) {
		uuid := "cb2fe2a7-ab9f-4f6d-9fd6-c7c209403cf0"
		retC := comment.Comment{
			Text:   "Test worknote 1",
			Entity: entity.NewEntity("incident", "f49d5fd5-8da4-4779-b5ba-32e78aa2c444"),
			UUID:   uuid,
			CreatedBy: &comment.UserInfo{
				UUID:           "8540d943-8ccd-4ff1-8a08-0c3aa338c58e",
				Name:           "Alice",
				Surname:        "Cooper",
				OrgName:        "a897a407-e41b-4b14-924a-39f5d5a8038f.kompitech.com",
				OrgDisplayName: "Kompitech",
			},
			CreatedAt: "2021-04-01T12:34:56+02:00",
		}

		assetType := comment.AssetTypeWorknote

		as := new(mocks.AuthServiceMock)
		as.On("Enforce", assetType.String(), auth.ReadAction, channelID, bearerToken).
			Return(true, nil)

		lister := new(mocks.ListingMock)
		lister.On("GetComment", uuid, channelID, assetType).
			Return(retC, nil)

		server := NewServer(Config{
			Addr:                    "service.url",
			AuthService:             as,
			Logger:                  logger,
			ListingService:          lister,
			ExternalLocationAddress: "http://service.url",
		})

		req := httptest.NewRequest("GET", "/worknotes/"+uuid, nil)
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
			"uuid":"cb2fe2a7-ab9f-4f6d-9fd6-c7c209403cf0",
			"text":"Test worknote 1",
			"entity":"incident:f49d5fd5-8da4-4779-b5ba-32e78aa2c444",
			"created_by":{
				"uuid":"8540d943-8ccd-4ff1-8a08-0c3aa338c58e",
				"name":"Alice",
				"surname":"Cooper",
				"org_name":"a897a407-e41b-4b14-924a-39f5d5a8038f.kompitech.com",
				"org_display_name":"Kompitech"
			},
			"created_at":"2021-04-01T12:34:56+02:00",
			"_links":{
				"self":{"href":"http://service.url/worknotes/cb2fe2a7-ab9f-4f6d-9fd6-c7c209403cf0"},
				"MarkWorknoteAsReadByUser":{"href":"http://service.url/worknotes/cb2fe2a7-ab9f-4f6d-9fd6-c7c209403cf0/read_by"}
			}
		}`
		assert.JSONEq(t, expectedJSON, string(b), "response does not match")
	})
}
