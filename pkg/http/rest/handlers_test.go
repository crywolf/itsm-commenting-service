package rest

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment"
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment/adding"
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment/listing"
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/entity"
	"github.com/KompiTech/itsm-commenting-service/pkg/repository/memory"
	"github.com/KompiTech/itsm-commenting-service/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddComment(t *testing.T) {
	logger := testutils.NewTestLogger()
	defer func() { _ = logger.Sync() }()

	rand := strings.NewReader("81aa058d-0b19-43e9-82ae-a7bca2457f10") // pseudo-random seed
	s := &memory.Storage{
		Clock: testutils.FixedClock{},
		Rand:  rand,
	}
	adder := adding.NewService(s)

	server := NewServer(Config{
		Addr:          "service.url",
		Logger:        logger,
		AddingService: adder,
	})

	payload := []byte(`{"entity":"incident:7e0d38d1-e5f5-4211-b2aa-3b142e4da80e", "text": "test with entity 1"}`)

	body := bytes.NewReader(payload)
	req := httptest.NewRequest("POST", "/comments", body)

	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)
	resp := w.Result()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	expectedLocation := "http://service.url/comments/38316161-3035-4864-ad30-6231392d3433"
	assert.Equal(t, expectedLocation, resp.Header.Get("Location"))
}

func TestGetComment(t *testing.T) {
	logger := testutils.NewTestLogger()
	defer func() { _ = logger.Sync() }()

	rand := strings.NewReader("81aa058d-0b19-43e9-82ae-a7bca2457f10") // pseudo-random seed
	s := &memory.Storage{
		Clock: testutils.FixedClock{},
		Rand:  rand,
	}

	c1 := comment.Comment{
		Text:   "Test comment 1",
		Entity: entity.NewEntity("incident", "f49d5fd5-8da4-4779-b5ba-32e78aa2c444"),
	}

	uuid, err := s.AddComment(c1)
	require.NoError(t, err)

	lister := listing.NewService(s)

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

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	expectedJSON := `{
		"uuid":"38316161-3035-4864-ad30-6231392d3433",
		"text":"Test comment 1",
		"entity":"incident:f49d5fd5-8da4-4779-b5ba-32e78aa2c444",
		"created_at":"2021-04-01T12:34:56+02:00"
	}`
	require.JSONEqf(t, expectedJSON, string(b), "response does not match")
}
