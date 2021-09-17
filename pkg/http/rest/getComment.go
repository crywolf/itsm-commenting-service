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
//	400: errorResponse400
//  401: errorResponse401
//  403: errorResponse403
//	404: errorResponse404

// swagger:route GET /worknotes/{uuid} worknotes GetWorknote
// Returns a single worknote from the repository
// responses:
//	200: commentResponse
//	400: errorResponse400
//  401: errorResponse401
//  403: errorResponse403
//	404: errorResponse404

// GetComment returns handler for getting single comment|worknote
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

		ctx := r.Context()

		asset, err := s.lister.GetComment(ctx, id, channelID, assetType)
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
