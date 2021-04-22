package rest

import (
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

	t.Run("when comment is being marked as read", func(t *testing.T) {
		us := new(mocks.UserServiceMock)
		us.On("UserBasicInfo", mock.AnythingOfType("*http.Request")).
			Return(mockUserData, nil)

		updater := new(mocks.UpdatingMock)
		updater.On(
			"MarkAsReadByUser",
			"7e0d38d1-e5f5-4211-b2aa-3b142e4da80e",
			mock.AnythingOfType("comment.ReadBy")).
			Return(false, nil)

		server := NewServer(Config{
			Addr:            "service.url",
			UserService:     us,
			Logger:          logger,
			UpdatingService: updater,
		})

		req := httptest.NewRequest("POST", "/comments/7e0d38d1-e5f5-4211-b2aa-3b142e4da80e/read_by", nil)

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
			mock.AnythingOfType("comment.ReadBy")).
			Return(true, nil)

		server := NewServer(Config{
			Addr:            "service.url",
			UserService:     us,
			Logger:          logger,
			UpdatingService: updater,
		})

		req := httptest.NewRequest("POST", "/comments/7e0d38d1-e5f5-4211-b2aa-3b142e4da80e/read_by", nil)

		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)
		resp := w.Result()

		assert.Equal(t, http.StatusNoContent, resp.StatusCode, "Status code")
		expectedLocation := "http://service.url/comments/7e0d38d1-e5f5-4211-b2aa-3b142e4da80e"
		assert.Equal(t, expectedLocation, resp.Header.Get("Location"), "Location header")
	})
}
