package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment"
	"github.com/KompiTech/itsm-commenting-service/pkg/http/rest/auth"
	"github.com/KompiTech/itsm-commenting-service/pkg/repository"
	"github.com/KompiTech/itsm-commenting-service/pkg/validation"
	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
)

// swagger:route POST /comments comments AddComment
// Creates a new comment
// responses:
//	201: createdResponse
//	400: errorResponse400
//	401: errorResponse401
//  403: errorResponse403
//	409: errorResponse409

// AddComment returns handler for POST /comments requests
func (s *Server) AddComment() func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	return s.addComment(assetTypeComment)
}

// swagger:route POST /worknotes worknotes AddWorknote
// Creates a new worknote
// responses:
//	201: createdResponse
//	400: errorResponse400
//	401: errorResponse401
//	403: errorResponse403
//	409: errorResponse409

// AddWorknote returns handler for POST /worknotes requests
func (s *Server) AddWorknote() func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	return s.addComment(assetTypeWorknote)
}

// addComment returns handler for POST requests
func (s *Server) addComment(assetType string) func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		s.logger.Info("AddComment handler called")

		if err := s.authorize("AddComment", assetType, auth.CreateAction, w, r); err != nil {
			return
		}

		defer func() { _ = r.Body.Close() }()
		payload, err := ioutil.ReadAll(r.Body)
		if err != nil {
			s.logger.Error("could not read request body", zap.Error(err))
			s.JSONError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var newComment comment.Comment

		err = s.payloadValidator.ValidatePayload(payload, "add_comment.yaml")
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

		err = json.Unmarshal(payload, &newComment)
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

		createdBy := &comment.UserInfo{
			UUID:           user.UUID,
			Name:           user.Name,
			Surname:        user.Surname,
			OrgName:        user.OrgName,
			OrgDisplayName: user.OrgDisplayName,
		}
		newComment.CreatedBy = createdBy

		ctx := r.Context()

		storedComment, err := s.adder.AddComment(ctx, newComment, channelID, assetType)
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

		commentBytes, err := json.Marshal(storedComment)
		if err != nil {
			s.logger.Error("AddComment handler failed", zap.Error(err))
			s.JSONError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		assetURI := fmt.Sprintf("%s/%s/%s", s.ExternalLocationAddress, pluralize(assetType), storedComment.UUID)

		w.Header().Set("Location", assetURI)
		w.WriteHeader(http.StatusCreated)
		w.Write(commentBytes)
	}
}

func pluralize(assetType string) string {
	return fmt.Sprintf("%ss", assetType)
}
