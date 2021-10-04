package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment"
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment/listing"
	"go.uber.org/zap"
)

// ActionType represents the name of the API action (e.g. GetComment)
type ActionType string

// Presenter provides REST responses
type Presenter interface {
	WriteGetResponse(r *http.Request, w http.ResponseWriter, comment comment.Comment, actionType ActionType)
	WriteListResponse(r *http.Request, w http.ResponseWriter, list listing.QueryResult, actionType ActionType)
	WriteError(w http.ResponseWriter, error string, code int)
}

// NewPresenter creates a presentation service
func NewPresenter(logger *zap.Logger, serverAddr string) Presenter {
	return &presenter{
		logger:     logger,
		serverAddr: serverAddr,
	}
}

type presenter struct {
	logger     *zap.Logger
	serverAddr string
}

type resourceContainer struct {
	comment.Comment
	Links []map[string]interface{} `json:"_links"`
}

type listContainer struct {
	listing.QueryResult
	Links []map[string]interface{} `json:"_links"`
}

func (s presenter) WriteGetResponse(_ *http.Request, w http.ResponseWriter, c comment.Comment, actionType ActionType) {
	resourceURI := fmt.Sprintf("%s%s", s.serverAddr, strings.ReplaceAll(string(actionType), "{uuid}", c.UUID))

	var readByRel string
	var readByAction ActionType
	switch actionType {
	case GetComment:
		readByRel = "MarkCommentAsReadByUser"
		readByAction = MarkCommentAsReadByUser
	case GetWorknote:
		readByRel = "MarkWorknoteAsReadByUser"
		readByAction = MarkWorknoteAsReadByUser
	}

	readByHref := fmt.Sprintf("%s%s", s.serverAddr, strings.ReplaceAll(string(readByAction), "{uuid}", c.UUID))

	links := []map[string]interface{}{
		{"rel": "self", "href": resourceURI},
		{"rel": readByRel, "href": readByHref},
	}

	s.encodeJSON(w, resourceContainer{Comment: c, Links: links})
}

func (s presenter) WriteListResponse(r *http.Request, w http.ResponseWriter, list listing.QueryResult, actionType ActionType) {
	delimiter := "?"
	resourceURI := fmt.Sprintf("%s%s", s.serverAddr, string(actionType))

	if r.URL.RawQuery != "" {
		delimiter = "&"
		resourceURI = fmt.Sprintf("%s?%s", resourceURI, r.URL.RawQuery)
	}

	links := []map[string]interface{}{
		{"rel": "self", "href": resourceURI},
	}

	if list.Bookmark != "" {
		bookmark := fmt.Sprintf("%sbookmark=%s", delimiter, list.Bookmark)
		next := map[string]interface{}{"rel": "next", "href": resourceURI + bookmark}
		links = append(links, next)
	}

	s.encodeJSON(w, listContainer{QueryResult: list, Links: links})
}

// SendError replies to the request with the specified error message and HTTP code.
// It does not otherwise end the request; the caller should ensure no further writes are done to 'w'.
// The error message should be plain text.
func (s presenter) WriteError(w http.ResponseWriter, error string, code int) {
	s.sendErrorJSON(w, error, code)
}

// sendErrorJSON replies to the request with the specified error message and HTTP code.
// It encodes error string as JSON object {"error":"error_string"} and sets correct header.
// It does not otherwise end the request; the caller should ensure no further writes are done to 'w'.
// The error message should be plain text.
func (s presenter) sendErrorJSON(w http.ResponseWriter, error string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	errorJSON, _ := json.Marshal(error)
	_, _ = fmt.Fprintf(w, `{"error":%s}`+"\n", errorJSON)
}

// encodeJSON encodes 'v' to JSON and writes it to the 'w'. Also sets correct Content-Type header.
// It does not otherwise end the request; the caller should ensure no further writes are done to 'w'.
func (s presenter) encodeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(v)
	if err != nil {
		eMsg := "could not encode JSON response"
		s.logger.Error(eMsg, zap.Error(err))
		s.WriteError(w, eMsg, http.StatusInternalServerError)
		return
	}
}
