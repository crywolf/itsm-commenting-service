package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/KompiTech/commenting-service/pkg/domain/comment"
	"github.com/KompiTech/commenting-service/pkg/repository"
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
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		id, err := s.adder.AddComment(newComment)
		if err != nil {
			s.logger.Error("AddComment handler failed", zap.Error(err))

			var httpError *repository.Error
			if errors.As(err, &httpError) {
				http.Error(w, err.Error(), httpError.StatusCode())
				return
			}

			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		assetURI := fmt.Sprintf("%s/comments/%s", s.Addr, id)

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
			var httpError *repository.Error
			if errors.As(err, &httpError) {
				http.Error(w, err.Error(), httpError.StatusCode())
				return
			}

			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(asset)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
