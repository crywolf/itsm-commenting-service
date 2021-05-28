package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/KompiTech/itsm-commenting-service/pkg/repository/couchdb"
	"github.com/KompiTech/itsm-commenting-service/testutils"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
)

// destroyTestDatabases deletes test databases
func destroyTestDatabases(storage *couchdb.DBStorage) {
	logger, _ := testutils.NewTestLogger()
	defer func() { _ = logger.Sync() }()

	c := storage.Client()
	if err := c.DestroyDB(context.TODO(), testutils.DatabaseName(testChannelID, "comment")); err != nil {
		logger.Warn("DestroyDB for comment", zap.Error(err))
	}
	if err := c.DestroyDB(context.TODO(), testutils.DatabaseName(testChannelID, "worknote")); err != nil {
		logger.Warn("DestroyDB for worknote", zap.Error(err))
	}
}

// createTestDatabases creates test databases for comments and worknotes
func createTestDatabases() {
	payload := []byte(`{"channel_id": "` + testChannelID + `"}`)
	body := bytes.NewReader(payload)
	req, err := http.NewRequest(http.MethodPost, server.URL+"/databases", body)
	Expect(err).To(BeNil())

	c := http.Client{}
	resp, err := c.Do(req)
	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Or(Equal(http.StatusCreated), Equal(http.StatusNoContent)), "createTestDatabases() failed: %s", resp.Status)
}

// createComment calls endpoint for comment creation and returns UUID of a newly created comment
func createComment(payload []byte) string {
	body := bytes.NewReader(payload)
	req, err := http.NewRequest(http.MethodPost, server.URL+"/comments", body)
	Expect(err).To(BeNil())
	req.Header.Set("grpc-metadata-space", testChannelID)
	req.Header.Set("authorization", bearerToken)

	c := http.Client{}
	resp, err := c.Do(req)
	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusCreated), "createComment() failed: %s", resp.Status)

	u, err := url.Parse(resp.Header.Get("Location"))
	Expect(err).To(BeNil())
	return strings.Split(u.Path, "/")[2]
}

// createWorknote calls the endpoint for worknote creation and returns UUID of a newly created worknote
func createWorknote(payload []byte) string {
	body := bytes.NewReader(payload)
	req, err := http.NewRequest(http.MethodPost, server.URL+"/worknotes", body)
	Expect(err).To(BeNil())
	req.Header.Set("grpc-metadata-space", testChannelID)
	req.Header.Set("authorization", bearerToken)

	c := http.Client{}
	resp, err := c.Do(req)
	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusCreated), "createWorknote() failed: %s", resp.Status)

	u, err := url.Parse(resp.Header.Get("Location"))
	Expect(err).To(BeNil())
	return strings.Split(u.Path, "/")[2]
}

// mapFromJSON converts JSON data into a map
func mapFromJSON(data []byte) map[string]interface{} {
	var result interface{}
	_ = json.Unmarshal(data, &result)
	return result.(map[string]interface{})
}
