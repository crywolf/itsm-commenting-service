package rest

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment/listing"
	"github.com/KompiTech/itsm-commenting-service/pkg/mocks"
	"github.com/KompiTech/itsm-commenting-service/pkg/repository/couchdb"
	"github.com/KompiTech/itsm-commenting-service/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestQueryCommentsHandler(t *testing.T) {
	logger := testutils.NewTestLogger()
	defer func() { _ = logger.Sync() }()

	t.Run("when some comments were found according to query", func(t *testing.T) {
		result := []map[string]interface{}{
			{
				"created_at": "2021-04-12T21:14:51+02:00",
				"text":       "test 1",
				"uuid":       "916c984f-e3fe-4638-8683-71f05501491f",
			},
			{
				"created_at": "2021-04-11T00:45:42+02:00",
				"text":       "test 4",
				"uuid":       "0ac5ebce-17e7-4edc-9552-fefe16e127fb",
			},
		}

		resultJSON, err := json.Marshal(result)
		if err != nil {
			t.Fatalf("could not marshall moc result: %v", err)
		}

		lister := new(mocks.ListingMock)
		lister.On("QueryComments", mock.AnythingOfType("map[string]interface {}")).
			Return(listing.QueryResult{Result: result}, nil)

		server := NewServer(Config{
			Addr:           "service.url",
			Logger:         logger,
			ListingService: lister,
		})

		payload := []byte(`{
			"selector":{"entity":"incident:7e0d38d1-e5f5-4211-b2aa-3b142e4da80e"},
			"sort":[{"created_at":"desc"}],
			"fields":["created_at","created_by","text","uuid","read_by"]
		}`)

		body := bytes.NewReader(payload)
		req := httptest.NewRequest("POST", "/comments/query", body)

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

		expectedJSON := `{"result":` + string(resultJSON) + `}`
		assert.JSONEq(t, expectedJSON, string(b), "response does not match")
	})

	t.Run("when bookmark is returned", func(t *testing.T) {
		result := []map[string]interface{}{
			{
				"created_at": "2021-04-12T21:14:51+02:00",
				"text":       "test 1",
				"uuid":       "916c984f-e3fe-4638-8683-71f05501491f",
			},
			{
				"created_at": "2021-04-11T00:45:42+02:00",
				"text":       "test 4",
				"uuid":       "0ac5ebce-17e7-4edc-9552-fefe16e127fb",
			},
			{
				"created_at": "2021-04-10T23:26:12+02:00",
				"text":       "test 3",
				"uuid":       "455e652a-5f5f-4c25-bb67-c0fb479fd5b1",
			},
		}

		resultJSON, err := json.Marshal(result)
		if err != nil {
			t.Fatalf("could not marshall moc result: %v", err)
		}

		lister := new(mocks.ListingMock)
		bookmark := "g1AAAAC2eJw1zjsOwjAQBNBVKKCi4hqL4vUncVpEGWioaJDttUVCCBKk4fY4CLrRSPM0AwAU14Jh_Zrcc7rF94UfoeN77rewO7bt_nACTkonXWk0IVhU1iu0ITB652vrapO0DzArq78y5P1iRpY_Y87YjZmO49TERCSjYZTGiwwyo0vCIguqtGJpKX4vbKgkgaVEUidhGkmNKs99_wEUVC69"
		lister.On("QueryComments", mock.AnythingOfType("map[string]interface {}")).
			Return(listing.QueryResult{
				Result:   result,
				Bookmark: bookmark,
			}, nil)

		server := NewServer(Config{
			Addr:           "service.url",
			Logger:         logger,
			ListingService: lister,
		})

		payload := []byte(`{
			"selector":{"entity":"incident:7e0d38d1-e5f5-4211-b2aa-3b142e4da80e"},
			"sort":[{"created_at":"desc"}],
			"fields":["created_at","created_by","text","uuid","read_by"]
		}`)

		body := bytes.NewReader(payload)
		req := httptest.NewRequest("POST", "/comments/query", body)

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
			"result":` + string(resultJSON) + `,
			"bookmark":"` + bookmark + `"
		}`
		assert.JSONEq(t, expectedJSON, string(b), "response does not match")
	})

	t.Run("when query is not a valid JSON and cannot be decoded", func(t *testing.T) {
		lister := new(mocks.ListingMock)
		server := NewServer(Config{
			Addr:           "service.url",
			Logger:         logger,
			ListingService: lister,
		})

		payload := []byte(`{this is not valid JSON at all}`)

		body := bytes.NewReader(payload)
		req := httptest.NewRequest("POST", "/comments/query", body)

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

		expectedJSON := `{"error":"could not decode JSON from request: invalid character 't' looking for beginning of object key string"}`
		assert.JSONEq(t, expectedJSON, string(b), "response does not match")
	})

	t.Run("when repository returns Bad Request error", func(t *testing.T) {
		lister := new(mocks.ListingMock)
		lister.On("QueryComments", mock.AnythingOfType("map[string]interface {}")).
			Return(listing.QueryResult{}, couchdb.ErrorBadRequest("index does not exist"))

		server := NewServer(Config{
			Addr:           "service.url",
			Logger:         logger,
			ListingService: lister,
		})

		payload := []byte(`{
			"selector":{"entity":"incident:7e0d38d1-e5f5-4211-b2aa-3b142e4da80e"},
			"sort":[{"created_at":"desc"}],
			"fields":["created_at","created_by","text","uuid","read_by"]
		}`)

		body := bytes.NewReader(payload)
		req := httptest.NewRequest("POST", "/comments/query", body)

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

		expectedJSON := `{"error":"index does not exist"}`
		assert.JSONEq(t, expectedJSON, string(b), "response does not match")
	})

	t.Run("when repository returns some other error", func(t *testing.T) {
		lister := new(mocks.ListingMock)
		lister.On("QueryComments", mock.AnythingOfType("map[string]interface {}")).
			Return(listing.QueryResult{}, errors.New("some error occurred"))

		server := NewServer(Config{
			Addr:           "service.url",
			Logger:         logger,
			ListingService: lister,
		})

		payload := []byte(`{
			"selector":{"entity":"incident:7e0d38d1-e5f5-4211-b2aa-3b142e4da80e"},
			"sort":[{"created_at":"desc"}],
			"fields":["created_at","created_by","text","uuid","read_by"]
		}`)

		body := bytes.NewReader(payload)
		req := httptest.NewRequest("POST", "/comments/query", body)

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
