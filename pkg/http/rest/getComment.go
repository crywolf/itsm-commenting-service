package rest

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/KompiTech/itsm-commenting-service/pkg/http/rest/auth"
	"github.com/KompiTech/itsm-commenting-service/pkg/repository"
	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
)

// swagger:route GET /comments/{uuid} comments GetComment
// Returns a single comment from the repository
// responses:
//	200: commentResponse
//	400: errorResponse
//  401: errorResponse
//  403: errorResponse
//	404: errorResponse

// GetComment returns handler for GET /comments/:id requests
func (s *Server) GetComment(assetType string) func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		s.logger.Info("GetComment handler called")

		if err := s.authorize("GetComment", assetType, auth.ReadAction, w, r); err != nil {
			return
		}

		id := params.ByName("id")
		if id == "" {
			eMsg := "malformed URL: missing resource ID param"
			s.logger.Warn("GetComment handler failed", zap.String("error", eMsg))
			s.JSONError(w, eMsg, http.StatusBadRequest)
			return
		}

		channelID, err := s.assertChannelID(w, r)
		if err != nil {
			return
		}

		asset, err := s.lister.GetComment(id, channelID, assetType)
		if err != nil {
			s.logger.Warn("GetComment handler failed", zap.Error(err))
			var httpError *repository.Error
			if errors.As(err, &httpError) {
				s.JSONError(w, err.Error(), httpError.StatusCode())
				return
			}

			s.logger.Error("GetComment handler failed", zap.Error(err))
			s.JSONError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(asset)
		if err != nil {
			eMsg := "could not encode JSON response"
			s.logger.Error(eMsg, zap.Error(err))
			s.JSONError(w, eMsg, http.StatusInternalServerError)
			return
		}
	}
}
