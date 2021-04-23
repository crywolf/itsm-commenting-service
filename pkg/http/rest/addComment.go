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

// AddComment returns handler for POST /comments requests
func (s *Server) AddComment() func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		s.logger.Info("AddComment handler called")

		decoder := json.NewDecoder(r.Body)

		var newComment comment.Comment
		err := decoder.Decode(&newComment)
		if err != nil {
			eMsg := "could not decode JSON from request"
			s.logger.Warn(eMsg, zap.Error(err))
			s.JSONError(w, fmt.Sprintf("%s: %s", eMsg, err.Error()), http.StatusBadRequest)
			return
		}

		channelID, err := s.assertChannelID(w, r)
		if err != nil {
			return
		}

		user, ok := s.UserInfoFromContext(r.Context())
		if !ok {
			eMsg := "could not get invoking user from context"
			s.logger.Error(eMsg)
			s.JSONError(w, eMsg, http.StatusInternalServerError)
			return
		}

		createdBy := &comment.CreatedBy{
			UUID:    user.UUID,
			Name:    user.Name,
			Surname: user.Surname,
		}
		newComment.CreatedBy = createdBy

		id, err := s.adder.AddComment(newComment, channelID)
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
