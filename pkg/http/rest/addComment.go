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
			// TODO test + JSON error
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		user, ok := UserFromContext(r.Context())
		if !ok {
			eMsg := "could not get invoking user from context"
			s.logger.Error(eMsg)
			s.JSONError(w, eMsg, http.StatusInternalServerError)
			return
		}

		createdBy := &comment.CreatedBy{
			UUID: user.UUID,
			Name: user.Name,
		}
		newComment.CreatedBy = createdBy

		id, err := s.adder.AddComment(newComment)
		if err != nil {
			var httpError *repository.Error
			if errors.As(err, &httpError) {
				s.logger.Warn("AddComment handler failed", zap.Error(err))
				s.JSONError(w, err.Error(), httpError.StatusCode())
				return
			}

			s.logger.Error("AddComment handler failed", zap.Error(err))
			s.JSONError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		URIschema := "http://"
		assetURI := fmt.Sprintf("%s%s/comments/%s", URIschema, s.Addr, id)

		w.Header().Set("Location", assetURI)
		w.WriteHeader(http.StatusCreated)
	}
}
