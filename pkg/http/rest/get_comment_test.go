package rest

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment"
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/entity"
	"github.com/KompiTech/itsm-commenting-service/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type ListingMock struct {
	mock.Mock
}

func (l *ListingMock) GetComment(id string) (comment.Comment, error) {
	args := l.Called(id)
	return args.Get(0).(comment.Comment), args.Error(1)
}

func TestGetCommentHandler(t *testing.T) {
	logger := testutils.NewTestLogger()
	defer func() { _ = logger.Sync() }()

	// we are adding comment first
	adder := new(AddingMock)
	adder.On("AddComment", mock.AnythingOfType("comment.Comment")).
		Return("cb2fe2a7-ab9f-4f6d-9fd6-c7c209403cf0", nil)

	c1 := comment.Comment{
		Text:   "Test comment 1",
		Entity: entity.NewEntity("incident", "f49d5fd5-8da4-4779-b5ba-32e78aa2c444"),
	}

	uuid, err := adder.AddComment(c1)
	require.NoError(t, err)

	retC := c1
	retC.UUID = "cb2fe2a7-ab9f-4f6d-9fd6-c7c209403cf0"
	retC.CreatedAt = "2021-04-01T12:34:56+02:00"

	// retrieving comment
	lister := new(ListingMock)
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
		"created_at":"2021-04-01T12:34:56+02:00"
	}`
	require.JSONEqf(t, expectedJSON, string(b), "response does not match")
}
