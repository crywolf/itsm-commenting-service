package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment"
	"github.com/KompiTech/itsm-commenting-service/pkg/repository"
	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
)

// AddComment returns a handler for POST /comments requests
func (s *Server) AddComment() func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		s.logger.Info("AddComment handler called")

		decoder := json.NewDecoder(r.Body)

		var newComment comment.Comment
		err := decoder.Decode(&newComment)
		if err != nil {
			s.logger.Warn("could not decode JSON from request", zap.Error(err))
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		id, err := s.adder.AddComment(newComment)
		if err != nil {
			var httpError *repository.Error
			if errors.As(err, &httpError) {
				s.logger.Warn("AddComment handler failed", zap.Error(err))
				http.Error(w, err.Error(), httpError.StatusCode())
				return
			}

			s.logger.Error("AddComment handler failed", zap.Error(err))
			msg := fmt.Sprintf("comment could not be created: %s", err.Error())
			http.Error(w, msg, http.StatusInternalServerError)
			return
		}

		URIschema := "http://"
		assetURI := fmt.Sprintf("%s%s/comments/%s", URIschema, s.Addr, id)

		w.Header().Set("Location", assetURI)
		w.WriteHeader(http.StatusCreated)
	}
}

// GetComment returns a handler for GET /comments requests
func (s *Server) GetComment() func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		s.logger.Info("GetComment handler called")

		id := params.ByName("id")

		asset, err := s.lister.GetComment(id)
		if err != nil {
			s.logger.Warn("GetComment handler failed", zap.Error(err))
			var httpError *repository.Error
			if errors.As(err, &httpError) {
				http.Error(w, err.Error(), httpError.StatusCode())
				return
			}

			s.logger.Error("GetComment handler failed", zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(asset)
		if err != nil {
			s.logger.Error("could not encode JSON response", zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
