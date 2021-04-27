package rest

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment"
	"github.com/KompiTech/itsm-commenting-service/pkg/repository"
	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
)

// MarkAsReadBy returns handler for POST /comments/:id/read_by requests
func (s *Server) MarkAsReadBy(assetType string) func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		s.logger.Info("MarkAsReadBy handler called")

		id := params.ByName("id")
		if id == "" {
			eMsg := "malformed URL: missing resource ID param"
			s.logger.Warn("MarkAsReadBy handler failed", zap.String("error", eMsg))
			s.JSONError(w, eMsg, http.StatusBadRequest)
			return
		}

		channelID, err := s.assertChannelID(w, r)
		if err != nil {
			return
		}

		user, ok := s.UserInfoFromContext(r.Context())
		if !ok {
			eMsg := "could not get invoking user info from context"
			s.logger.Error(eMsg)
			s.JSONError(w, eMsg, http.StatusInternalServerError)
			return
		}

		readBy := comment.ReadBy{
			Time: time.Now().Format(time.RFC3339),
			User: comment.UserInfo{
				UUID:           user.UUID,
				Name:           user.Name,
				Surname:        user.Surname,
				OrgName:        user.OrgName,
				OrgDisplayName: user.OrgDisplayName,
			},
		}

		alreadyMarked, err := s.updater.MarkAsReadByUser(id, readBy, channelID, assetType)
		if err != nil {
			var httpError *repository.Error
			if errors.As(err, &httpError) {
				s.logger.Warn("MarkAsReadBy handler failed", zap.Error(err))
				s.JSONError(w, err.Error(), httpError.StatusCode())
				return
			}

			s.logger.Error("MarkAsReadBy handler failed", zap.Error(err))
			s.JSONError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		URIschema := "http://"
		assetURI := fmt.Sprintf("%s%s/%s/%s", URIschema, s.Addr, pluralize(assetType), id)

		w.Header().Set("Location", assetURI)

		if alreadyMarked {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}
