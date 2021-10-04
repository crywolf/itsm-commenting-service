package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment"
	"github.com/KompiTech/itsm-commenting-service/pkg/http/rest/auth"
	"github.com/KompiTech/itsm-commenting-service/pkg/repository"
	"github.com/julienschmidt/httprouter"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
)

// ListComments route
const ListComments ActionType = "/comments"

// swagger:route GET /comments comments ListComments
// Returns a list of comments from the repository filtered by some parameters
// responses:
//	200: commentsListResponse
//	400: errorResponse400
//  401: errorResponse401
//  403: errorResponse403

// ListWorknotes route
const ListWorknotes ActionType = "/worknotes"

// swagger:route GET /worknotes worknotes ListWorknotes
// Returns a list of worknotes from the repository filtered by some parameters
// responses:
//	200: commentsListResponse
//	400: errorResponse400
//  401: errorResponse401
//  403: errorResponse403

// QueryComments returns handler for listing or querying comments|worknotes
func (s *Server) QueryComments(assetType comment.AssetType) func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		s.logger.Info("QueryComments handler called")
		span, ctx := opentracing.StartSpanFromContext(r.Context(), "itsm-commenting-service-query")
		defer span.Finish()

		r = r.WithContext(ctx)

		if err := s.authorize("QueryComments", assetType.String(), auth.ReadAction, w, r); err != nil {
			return
		}

		var query = map[string]interface{}{}

		queryValues := r.URL.Query()
		queryParam := queryValues.Get("query")
		if queryParam != "" {
			JSONquery, err := url.QueryUnescape(queryValues.Get("query"))
			if err != nil {
				msg := "could not unescape JSON query from request"
				s.logger.Warn(msg, zap.Error(err))
				s.presenter.WriteError(w, fmt.Sprintf("%s: %v", msg, err.Error()), http.StatusBadRequest)
				return
			}

			decoder := json.NewDecoder(strings.NewReader(JSONquery))
			err = decoder.Decode(&query)
			if err != nil {
				msg := "could not decode JSON query from request"
				s.logger.Warn(msg, zap.Error(err))
				s.presenter.WriteError(w, fmt.Sprintf("%s: %v", msg, err.Error()), http.StatusBadRequest)
				return
			}
		}

		// no query param => we create our query
		if len(query) == 0 {
			entity := queryValues.Get("entity")
			if entity != "" {
				// list all comments that belongs to one entity
				query["selector"] = map[string]string{"entity": entity}
			} else {
				// list all comments
				query["selector"] = map[string]interface{}{"_id": map[string]interface{}{"$gt": nil}}
			}
			limit := queryValues.Get("limit")
			if limit != "" {
				l, _ := strconv.ParseFloat(limit, 64)
				query["limit"] = l
			}
			bookmark := queryValues.Get("bookmark")
			if bookmark != "" {
				query["bookmark"] = bookmark
			}

			query["sort"] = []map[string]string{{"created_at": "desc"}}
			query["fields"] = []string{"created_at", "created_by", "text", "entity", "uuid", "read_by"}
		}

		channelID, err := s.assertChannelID(w, r)
		if err != nil {
			return
		}

		qResult, err := s.lister.QueryComments(ctx, query, channelID, assetType)
		if err != nil {
			var httpError *repository.Error
			if errors.As(err, &httpError) {
				s.logger.Error("Repository error", zap.Error(err))
				s.presenter.WriteError(w, err.Error(), httpError.StatusCode())
				return
			}

			s.logger.Error("GetComment handler failed", zap.Error(err))
			s.presenter.WriteError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		s.presenter.WriteListResponse(r, w, qResult, assetType)
	}
}
