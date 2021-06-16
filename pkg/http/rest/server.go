package rest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment/adding"
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment/listing"
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment/updating"
	"github.com/KompiTech/itsm-commenting-service/pkg/http/rest/auth"
	"github.com/KompiTech/itsm-commenting-service/pkg/http/rest/validation"
	"github.com/KompiTech/itsm-commenting-service/pkg/repository"
	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
)

// Server is a http.Handler with dependencies
type Server struct {
	Addr              string
	URISchema         string
	router            *httprouter.Router
	logger            *zap.Logger
	authService       auth.Service
	userService       UserService
	adder             adding.Service
	lister            listing.Service
	updater           updating.Service
	repositoryService repository.Service
	payloadValidator  validation.PayloadValidator
}

// Config contains server configuration and dependencies
type Config struct {
	Addr              string
	URISchema         string
	Logger            *zap.Logger
	AuthService       auth.Service
	UserService       UserService
	AddingService     adding.Service
	ListingService    listing.Service
	UpdatingService   updating.Service
	RepositoryService repository.Service
	PayloadValidator  validation.PayloadValidator
}

// NewServer creates new server with the necessary dependencies
func NewServer(cfg Config) *Server {
	r := httprouter.New()

	URISchema := "http://"
	if cfg.URISchema != "" {
		URISchema = cfg.URISchema
	}

	s := &Server{
		Addr:              cfg.Addr,
		URISchema:         URISchema,
		router:            r,
		logger:            cfg.Logger,
		authService:       cfg.AuthService,
		userService:       cfg.UserService,
		adder:             cfg.AddingService,
		lister:            cfg.ListingService,
		updater:           cfg.UpdatingService,
		repositoryService: cfg.RepositoryService,
		payloadValidator:  cfg.PayloadValidator,
	}
	s.routes()

	return s
}

// ServeHTTP makes the server implement the http.Handler interface
func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	channelID := r.Header.Get("grpc-metadata-space")
	ctx := context.WithValue(r.Context(), channelIDKey, channelID)

	authToken := r.Header.Get("authorization")
	ctx = context.WithValue(ctx, authKey, authToken)

	s.router.ServeHTTP(w, r.WithContext(ctx))
}

// JSONError replies to the request with the specified error message and HTTP code.
// It encode error string as JSON object {"error":"error_string"} and sets correct header.
// It does not otherwise end the request; the caller should ensure no further  writes are done to w.
// The error message should be plain text.
func (s Server) JSONError(w http.ResponseWriter, error string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	errorJSON, _ := json.Marshal(error)
	_, _ = fmt.Fprintf(w, `{"error":%s}`+"\n", errorJSON)
}

type channelIDType int

var channelIDKey channelIDType

// channelIDFromContext returns channelID stored in ctx, if any.
func channelIDFromContext(ctx context.Context) (string, bool) {
	ch, ok := ctx.Value(channelIDKey).(string)
	return ch, ok
}

// assertChannelID writes error message to response and returns error if channelID cannot be determined,
// otherwise it returns channelID
func (s Server) assertChannelID(w http.ResponseWriter, r *http.Request) (string, error) {
	channelID, ok := channelIDFromContext(r.Context())
	if !ok {
		eMsg := "could not get channel ID from context"
		s.logger.Error(eMsg)
		s.JSONError(w, eMsg, http.StatusInternalServerError)
		return "", errors.New(eMsg)
	}

	if channelID == "" {
		eMsg := "empty channel ID in context"
		s.logger.Error(eMsg)
		s.JSONError(w, "'grpc-metadata-space' header missing or invalid", http.StatusUnauthorized)
		return "", errors.New(eMsg)
	}

	return channelID, nil
}

type authType int

var authKey authType

// authTokenFromContext returns authorization token stored in ctx, if any.
func authTokenFromContext(ctx context.Context) (string, bool) {
	ch, ok := ctx.Value(authKey).(string)
	return ch, ok
}

// assertAuthToken writes error message to response and returns error if authorization token cannot be determined,
// otherwise it returns authorization token
func (s Server) assertAuthToken(w http.ResponseWriter, r *http.Request) (string, error) {
	authToken, ok := authTokenFromContext(r.Context())
	if !ok {
		eMsg := "could not get authorization token from context"
		s.logger.Error(eMsg)
		s.JSONError(w, eMsg, http.StatusInternalServerError)
		return "", errors.New(eMsg)
	}

	if authToken == "" {
		eMsg := "empty authorization token in context"
		s.logger.Error(eMsg)
		s.JSONError(w, "'authorization' header missing or invalid", http.StatusUnauthorized)
		return "", errors.New(eMsg)
	}

	return authToken, nil
}

// authorize checks if user is authorized to perform action on asset,
// otherwise it writes error message to response and returns error to notify calling handler to stop execution
func (s *Server) authorize(handlerName, assetType string, action auth.Action, w http.ResponseWriter, r *http.Request) error {
	authToken, err := s.assertAuthToken(w, r)
	if err != nil {
		return err
	}

	channelID, err := s.assertChannelID(w, r)
	if err != nil {
		return err
	}

	if onBehalf := r.Header.Get("on_behalf"); onBehalf != "" {
		action, err = action.OnBehalf()
		if err != nil {
			eMsg := fmt.Sprintf("authorization failed: %v", err)
			s.JSONError(w, eMsg, http.StatusInternalServerError)
			return err
		}
	}

	authorized, err := s.authService.Enforce(assetType, action, channelID, authToken)
	if err != nil {
		s.logger.Error(fmt.Sprintf("%s handler failed", handlerName), zap.Error(err))
		s.JSONError(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	if !authorized {
		eMsg := fmt.Sprintf("Forbidden (%s, %s)", assetType, action)
		s.logger.Warn(fmt.Sprintf("%s handler failed", handlerName), zap.String("msg", eMsg))
		s.JSONError(w, eMsg, http.StatusForbidden)
		return errors.New(eMsg)
	}

	return nil
}
