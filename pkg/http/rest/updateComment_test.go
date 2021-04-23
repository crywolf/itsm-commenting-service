package rest

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KompiTech/itsm-commenting-service/pkg/domain/user"
	"github.com/KompiTech/itsm-commenting-service/pkg/mocks"
	"github.com/KompiTech/itsm-commenting-service/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMarkAsReadByHandler(t *testing.T) {
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

		server := NewServer(Config{
			Addr:        "service.url",
			UserService: us,
			Logger:      logger,
		})

		req := httptest.NewRequest("POST", "/comments/7e0d38d1-e5f5-4211-b2aa-3b142e4da80e/read_by", nil)

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

	t.Run("when comment is being marked as read", func(t *testing.T) {
		us := new(mocks.UserServiceMock)
		us.On("UserBasicInfo", mock.AnythingOfType("*http.Request")).
			Return(mockUserData, nil)

		updater := new(mocks.UpdatingMock)
		updater.On(
			"MarkAsReadByUser",
			"7e0d38d1-e5f5-4211-b2aa-3b142e4da80e",
			mock.AnythingOfType("comment.ReadBy"), channelID).
			Return(false, nil)

		server := NewServer(Config{
			Addr:            "service.url",
			UserService:     us,
			Logger:          logger,
			UpdatingService: updater,
		})

		req := httptest.NewRequest("POST", "/comments/7e0d38d1-e5f5-4211-b2aa-3b142e4da80e/read_by", nil)
		req.Header.Set("grpc-metadata-space", channelID)

		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)
		resp := w.Result()

		assert.Equal(t, http.StatusCreated, resp.StatusCode, "Status code")
		expectedLocation := "http://service.url/comments/7e0d38d1-e5f5-4211-b2aa-3b142e4da80e"
		assert.Equal(t, expectedLocation, resp.Header.Get("Location"), "Location header")
	})

	t.Run("when comment is being marked as read twice by the same user", func(t *testing.T) {
		us := new(mocks.UserServiceMock)
		us.On("UserBasicInfo", mock.AnythingOfType("*http.Request")).
			Return(mockUserData, nil)

		updater := new(mocks.UpdatingMock)
		updater.On(
			"MarkAsReadByUser",
			"7e0d38d1-e5f5-4211-b2aa-3b142e4da80e",
			mock.AnythingOfType("comment.ReadBy"), channelID).
			Return(true, nil)

		server := NewServer(Config{
			Addr:            "service.url",
			UserService:     us,
			Logger:          logger,
			UpdatingService: updater,
		})

		req := httptest.NewRequest("POST", "/comments/7e0d38d1-e5f5-4211-b2aa-3b142e4da80e/read_by", nil)
		req.Header.Set("grpc-metadata-space", channelID)

		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)
		resp := w.Result()

		assert.Equal(t, http.StatusNoContent, resp.StatusCode, "Status code")
		expectedLocation := "http://service.url/comments/7e0d38d1-e5f5-4211-b2aa-3b142e4da80e"
		assert.Equal(t, expectedLocation, resp.Header.Get("Location"), "Location header")
	})
}
