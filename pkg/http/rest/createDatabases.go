package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/KompiTech/itsm-commenting-service/pkg/repository"
	"github.com/KompiTech/itsm-commenting-service/pkg/validation"
	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
)

// swagger:route POST /databases databases CreateDatabases
// Creates new databases for channel; if databases already exist it just returns 204 No Content
//
// responses:
//	201: databasesCreatedResponse
//	204: databasesNoContentResponse
//	400: errorResponse

// CreateDatabases returns handler for POST /databases requests
func (s *Server) CreateDatabases() func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	type requestBody struct {
		ChannelID string `json:"channel_id"`
	}

	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		s.logger.Info("CreateDatabases handler called")

		payload, err := ioutil.ReadAll(r.Body)
		if err != nil {
			s.logger.Error("could not read request body", zap.Error(err))
			s.JSONError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		defer func() { _ = r.Body.Close() }()

		err = s.payloadValidator.ValidatePayload(payload, "create_databases.yaml")
		if err != nil {
			var errGeneral *validation.ErrGeneral
			if errors.As(err, &errGeneral) {
				s.logger.Error("payload validation", zap.Error(err))
				s.JSONError(w, err.Error(), http.StatusInternalServerError)
				return
			}

			s.logger.Warn("invalid payload", zap.Error(err))
			s.JSONError(w, err.Error(), http.StatusBadRequest)
			return
		}

		var request requestBody
		err = json.Unmarshal(payload, &request)
		if err != nil {
			eMsg := "could not decode JSON from request"
			s.logger.Warn(eMsg, zap.Error(err))
			s.JSONError(w, fmt.Sprintf("%s: %s", eMsg, err.Error()), http.StatusBadRequest)
			return
		}

		assetTypes := [2]string{"comment", "worknote"}

		bothExisted := true

		for _, assetType := range assetTypes {
			alreadyExisted, err := s.repositoryService.CreateDatabase(request.ChannelID, assetType)
			if err != nil {
				var httpError *repository.Error
				if errors.As(err, &httpError) {
					s.logger.Warn("CreateDatabases handler failed", zap.Error(err), zap.String("assetType", assetType))
					s.JSONError(w, err.Error(), httpError.StatusCode())
					return
				}

				s.logger.Warn("CreateDatabases handler failed", zap.Error(err), zap.String("assetType", assetType))
				s.JSONError(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if !alreadyExisted {
				bothExisted = false
			}
		}

		if bothExisted {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = fmt.Fprintln(w, `{"message":"databases were successfully created"}`)
	}
}
