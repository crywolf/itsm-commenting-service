package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment"
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment/listing"
	"github.com/KompiTech/itsm-commenting-service/pkg/http/rest/hypermedia"
	"go.uber.org/zap"
)

// ActionType represents the name of the API action (e.g. GetComment)
type ActionType string

func (a ActionType) String() string {
	return string(a)
}

// Presenter provides REST responses
type Presenter interface {
	WriteGetResponse(r *http.Request, w http.ResponseWriter, comment comment.Comment, assetType comment.AssetType)
	WriteListResponse(r *http.Request, w http.ResponseWriter, list listing.QueryResult, assetType comment.AssetType)
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

func (p presenter) WriteGetResponse(_ *http.Request, w http.ResponseWriter, c comment.Comment, assetType comment.AssetType) {
	var action ActionType
	switch assetType {
	case comment.AssetTypeComment:
		action = GetComment
	case comment.AssetTypeWorknote:
		action = GetWorknote
	}

	resourceURI := fmt.Sprintf("%s%s", p.serverAddr, strings.ReplaceAll(action.String(), "{uuid}", c.UUID))

	links := map[string]interface{}{
		"self": map[string]string{"href": resourceURI},
	}

	allowedLinks := hypermedia.AllowedLinksForComment(c, assetType)
	for _, linkName := range allowedLinks {
		action, err := p.mapLinkNameToAction(linkName)
		if err != nil {
			p.WriteError(w, err.Error(), http.StatusInternalServerError)
		}

		href := fmt.Sprintf("%s%s", p.serverAddr, strings.ReplaceAll(action.String(), "{uuid}", c.UUID))

		links[linkName] = map[string]string{
			"href": href,
		}
	}

	p.encodeJSON(w, resourceContainer{Comment: c, Links: links})
}

func (p presenter) WriteListResponse(r *http.Request, w http.ResponseWriter, list listing.QueryResult, assetType comment.AssetType) {
	var action ActionType
	switch assetType {
	case comment.AssetTypeComment:
		action = ListComments
	case comment.AssetTypeWorknote:
		action = ListWorknotes
	}
	resourceURI := fmt.Sprintf("%s%s", p.serverAddr, action)

	delimiter := "?"

	if r.URL.RawQuery != "" {
		delimiter = "&"
		resourceURI = fmt.Sprintf("%s?%s", resourceURI, r.URL.RawQuery)
	}

	links := map[string]interface{}{
		"self": map[string]string{"href": resourceURI},
	}

	if list.Bookmark != "" {
		bookmark := fmt.Sprintf("%sbookmark=%s", delimiter, list.Bookmark)
		links["next"] = map[string]string{
			"href": resourceURI + bookmark,
		}
	}

	p.encodeJSON(w, listContainer{QueryResult: list, Links: links})
}

// WriteError replies to the request with the specified error message and HTTP code.
// It does not otherwise end the request; the caller should ensure no further writes are done to 'w'.
// The error message should be plain text.
func (p presenter) WriteError(w http.ResponseWriter, error string, code int) {
	p.sendErrorJSON(w, error, code)
}

// sendErrorJSON replies to the request with the specified error message and HTTP code.
// It encodes error string as JSON object {"error":"error_string"} and sets correct header.
// It does not otherwise end the request; the caller should ensure no further writes are done to 'w'.
// The error message should be plain text.
func (p presenter) sendErrorJSON(w http.ResponseWriter, error string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	errorJSON, _ := json.Marshal(error)
	_, _ = fmt.Fprintf(w, `{"error":%s}`+"\n", errorJSON)
}

// encodeJSON encodes 'v' to JSON and writes it to the 'w'. Also sets correct Content-Type header.
// It does not otherwise end the request; the caller should ensure no further writes are done to 'w'.
func (p presenter) encodeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(v)
	if err != nil {
		eMsg := "could not encode JSON response"
		p.logger.Error(eMsg, zap.Error(err))
		p.WriteError(w, eMsg, http.StatusInternalServerError)
		return
	}
}

func (p presenter) mapLinkNameToAction(name string) (ActionType, error) {
	m := map[string]ActionType{
		"MarkCommentAsReadByUser":  MarkCommentAsReadByUser,
		"MarkWorknoteAsReadByUser": MarkWorknoteAsReadByUser,
	}

	action, ok := m[name]
	if !ok {
		return "", fmt.Errorf("mapping link name to action failed, name %s is not defined", name)
	}

	return action, nil
}

type resourceContainer struct {
	comment.Comment
	Links map[string]interface{} `json:"_links"`
}

type listContainer struct {
	listing.QueryResult
	Links map[string]interface{} `json:"_links"`
}
