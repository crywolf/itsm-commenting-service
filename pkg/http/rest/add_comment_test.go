package rest

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment"
	"github.com/KompiTech/itsm-commenting-service/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type AddingMock struct {
	mock.Mock
}

func (a *AddingMock) AddComment(c comment.Comment) (string, error) {
	args := a.Called(c)
	return args.String(0), args.Error(1)
}

func TestAddCommentHandler(t *testing.T) {
	logger := testutils.NewTestLogger()
	defer func() { _ = logger.Sync() }()

	adder := new(AddingMock)
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
}
