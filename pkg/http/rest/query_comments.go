package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/KompiTech/itsm-commenting-service/pkg/repository"
	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
)

// QueryComments returns a handler for POST /comments/query requests
func (s *Server) QueryComments() func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		s.logger.Info("ListComments handler called")

		decoder := json.NewDecoder(r.Body)

		var request map[string]interface{}
		err := decoder.Decode(&request)
		if err != nil {
			msg := "could not decode JSON from request"
			s.logger.Warn(msg, zap.Error(err))
			s.JSONError(w, fmt.Sprintf("%s: %v", msg, err.Error()), http.StatusBadRequest)
			return
		}

		qResult, err := s.lister.QueryComments(request)
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
