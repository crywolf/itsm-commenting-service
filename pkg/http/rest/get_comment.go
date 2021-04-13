package rest

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/KompiTech/itsm-commenting-service/pkg/repository"
	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
)

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
