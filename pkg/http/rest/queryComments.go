package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/KompiTech/itsm-commenting-service/pkg/repository"
	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
)

// QueryComments returns a handler for POST /comments/query requests
func (s *Server) QueryComments() func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		s.logger.Info("ListComments handler called")

		var query = map[string]interface{}{}

		queryValues := r.URL.Query()
		queryParam := queryValues.Get("query")
		if queryParam != "" {
			JSONquery, err := url.QueryUnescape(queryValues.Get("query"))
			if err != nil {
				msg := "could not unescape JSON query from request"
				s.logger.Warn(msg, zap.Error(err))
				s.JSONError(w, fmt.Sprintf("%s: %v", msg, err.Error()), http.StatusBadRequest)
				return
			}

			decoder := json.NewDecoder(strings.NewReader(JSONquery))
			err = decoder.Decode(&query)
			if err != nil {
				msg := "could not decode JSON query from request"
				s.logger.Warn(msg, zap.Error(err))
				s.JSONError(w, fmt.Sprintf("%s: %v", msg, err.Error()), http.StatusBadRequest)
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
			query["fields"] = []string{"created_at", "created_by", "text", "uuid", "read_by"}
		}

		qResult, err := s.lister.QueryComments(query)
		if err != nil {
			var httpError *repository.Error
			if errors.As(err, &httpError) {
				s.logger.Error("Repository error", zap.Error(err))
				s.JSONError(w, err.Error(), httpError.StatusCode())
				return
			}

			s.logger.Error("GetComment handler failed", zap.Error(err))
			s.JSONError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(qResult)
		if err != nil {
			s.logger.Error("could not encode JSON response", zap.Error(err))
			s.JSONError(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
