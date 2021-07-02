package rest

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KompiTech/itsm-commenting-service/pkg/domain/user"
	"github.com/KompiTech/itsm-commenting-service/pkg/http/rest/auth"
	"github.com/KompiTech/itsm-commenting-service/pkg/mocks"
	"github.com/KompiTech/itsm-commenting-service/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMarkAsReadByHandler(t *testing.T) {
	logger, _ := testutils.NewTestLogger()
	defer func() { _ = logger.Sync() }()

	mockUserData := user.BasicInfo{
		UUID: "2af4f493-0bd5-4513-b440-6cbb465feadb",
		Name: "Some test user 1",
	}

	channelID := "e27ddcd0-0e1f-4bc5-93df-f6f04155beec"
	bearerToken := "some valid Bearer token"

	t.Run("when channelID is not set (ie. grpc-metadata-space header is missing)", func(t *testing.T) {
		as := new(mocks.AuthServiceMock)

		us := new(mocks.UserServiceMock)
		us.On("UserBasicInfo", mock.AnythingOfType("*http.Request")).
			Return(mockUserData, nil)

		server := NewServer(Config{
			Addr:        "service.url",
			Logger:      logger,
			AuthService: as,
			UserService: us,
		})

		req := httptest.NewRequest("POST", "/comments/7e0d38d1-e5f5-4211-b2aa-3b142e4da80e/read_by", nil)
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

	t.Run("when user is not authorized to READ the comment", func(t *testing.T) {
		as := new(mocks.AuthServiceMock)
		as.On("Enforce", "comment", auth.ReadAction, channelID, bearerToken).
			Return(false, nil)

		us := new(mocks.UserServiceMock)
		us.On("UserBasicInfo", mock.AnythingOfType("*http.Request")).
			Return(mockUserData, nil)

		server := NewServer(Config{
			Addr:        "service.url",
			Logger:      logger,
			AuthService: as,
			UserService: us,
			ExternalLocationAddress: "http://service.url",
		})

		req := httptest.NewRequest("POST", "/comments/7e0d38d1-e5f5-4211-b2aa-3b142e4da80e/read_by", nil)
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

		assert.Equal(t, http.StatusForbidden, resp.StatusCode, "Status code")
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"), "Content-Type header")

		expectedJSON := `{"error":"Authorization failed, action forbidden (comment, read)"}`
		assert.JSONEq(t, expectedJSON, string(b), "response does not match")
	})

	t.Run("when comment is being marked as read", func(t *testing.T) {
		as := new(mocks.AuthServiceMock)
		as.On("Enforce", "comment", auth.ReadAction, channelID, bearerToken).
			Return(true, nil)

		us := new(mocks.UserServiceMock)
		us.On("UserBasicInfo", mock.AnythingOfType("*http.Request")).
			Return(mockUserData, nil)

		updater := new(mocks.UpdatingMock)
		assetType := "comment"
		updater.On(
			"MarkAsReadByUser",
			"7e0d38d1-e5f5-4211-b2aa-3b142e4da80e",
			mock.AnythingOfType("comment.ReadBy"), channelID, assetType).
			Return(false, nil)

		server := NewServer(Config{
			Addr:            "service.url",
			Logger:          logger,
			AuthService:     as,
			UserService:     us,
			UpdatingService: updater,
			ExternalLocationAddress: "http://service.url",
		})

		req := httptest.NewRequest("POST", "/comments/7e0d38d1-e5f5-4211-b2aa-3b142e4da80e/read_by", nil)
		req.Header.Set("grpc-metadata-space", channelID)
		req.Header.Set("authorization", bearerToken)

		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)
		resp := w.Result()

		assert.Equal(t, http.StatusCreated, resp.StatusCode, "Status code")
		expectedLocation := "http://service.url/comments/7e0d38d1-e5f5-4211-b2aa-3b142e4da80e"
		assert.Equal(t, expectedLocation, resp.Header.Get("Location"), "Location header")
	})

	t.Run("when comment is being marked as read twice by the same user", func(t *testing.T) {
		as := new(mocks.AuthServiceMock)
		as.On("Enforce", "comment", auth.ReadAction, channelID, bearerToken).
			Return(true, nil)

		us := new(mocks.UserServiceMock)
		us.On("UserBasicInfo", mock.AnythingOfType("*http.Request")).
			Return(mockUserData, nil)

		updater := new(mocks.UpdatingMock)
		assetType := "comment"
		updater.On(
			"MarkAsReadByUser",
			"7e0d38d1-e5f5-4211-b2aa-3b142e4da80e",
			mock.AnythingOfType("comment.ReadBy"), channelID, assetType).
			Return(true, nil)

		server := NewServer(Config{
			Addr:            "service.url",
			Logger:          logger,
			AuthService:     as,
			UserService:     us,
			UpdatingService: updater,
			ExternalLocationAddress: "http://service.url",
		})

		req := httptest.NewRequest("POST", "/comments/7e0d38d1-e5f5-4211-b2aa-3b142e4da80e/read_by", nil)
		req.Header.Set("grpc-metadata-space", channelID)
		req.Header.Set("authorization", bearerToken)

		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)
		resp := w.Result()

		assert.Equal(t, http.StatusNoContent, resp.StatusCode, "Status code")
		expectedLocation := "http://service.url/comments/7e0d38d1-e5f5-4211-b2aa-3b142e4da80e"
		assert.Equal(t, expectedLocation, resp.Header.Get("Location"), "Location header")
	})

	// worknote
	t.Run("when worknote is being marked as read", func(t *testing.T) {
		assetType := "worknote"
		as := new(mocks.AuthServiceMock)
		as.On("Enforce", assetType, auth.ReadAction, channelID, bearerToken).
			Return(true, nil)

		us := new(mocks.UserServiceMock)
		us.On("UserBasicInfo", mock.AnythingOfType("*http.Request")).
			Return(mockUserData, nil)

		updater := new(mocks.UpdatingMock)
		updater.On(
			"MarkAsReadByUser",
			"7e0d38d1-e5f5-4211-b2aa-3b142e4da80e",
			mock.AnythingOfType("comment.ReadBy"), channelID, assetType).
			Return(false, nil)

		server := NewServer(Config{
			Addr:            "service.url",
			Logger:          logger,
			AuthService:     as,
			UserService:     us,
			UpdatingService: updater,
			ExternalLocationAddress: "http://service.url",
		})

		req := httptest.NewRequest("POST", "/worknotes/7e0d38d1-e5f5-4211-b2aa-3b142e4da80e/read_by", nil)
		req.Header.Set("grpc-metadata-space", channelID)
		req.Header.Set("authorization", bearerToken)

		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)
		resp := w.Result()

		assert.Equal(t, http.StatusCreated, resp.StatusCode, "Status code")
		expectedLocation := "http://service.url/worknotes/7e0d38d1-e5f5-4211-b2aa-3b142e4da80e"
		assert.Equal(t, expectedLocation, resp.Header.Get("Location"), "Location header")
	})
}
