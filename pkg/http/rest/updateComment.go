package rest

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment"
	"github.com/KompiTech/itsm-commenting-service/pkg/http/rest/auth"
	"github.com/KompiTech/itsm-commenting-service/pkg/repository"
	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
)

// swagger:route POST /comments/{uuid}/read_by comments MarkCommentAsReadByUser
// Marks specified comment as read by user
// responses:
//	201: createdResponse
//	204: noContentResponse
//	400: errorResponse400
//	401: errorResponse401
//  403: errorResponse403
//	404: errorResponse404

// swagger:route POST /worknotes/{uuid}/read_by worknotes MarkWorknoteAsReadByUser
// Marks specified worknote as read by user
// responses:
//	201: createdResponse
//	204: noContentResponse
//	400: errorResponse400
//	401: errorResponse401
//  403: errorResponse403
//	404: errorResponse404

// MarkCommentAsReadBy returns handler for marking comment|worknote as read by user
func (s *Server) MarkCommentAsReadBy(assetType string) func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		s.logger.Info("MarkAsReadBy handler called")

		// use can update comment if he is allowed to read it!
		if err := s.authorize("MarkAsReadBy", assetType, auth.ReadAction, w, r); err != nil {
			return
		}

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

		alreadyMarked, err := s.updater.MarkAsReadByUser(context.Background(), id, readBy, channelID, assetType)
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

		assetURI := fmt.Sprintf("%s/%s/%s", s.ExternalLocationAddress, pluralize(assetType), id)

		w.Header().Set("Location", assetURI)

		if alreadyMarked {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}
