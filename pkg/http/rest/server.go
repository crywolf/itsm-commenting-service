package rest

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment/adding"
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment/listing"
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment/updating"
	"github.com/KompiTech/itsm-commenting-service/pkg/http/rest/auth"
	"github.com/KompiTech/itsm-commenting-service/pkg/http/rest/usersvc"
	"github.com/KompiTech/itsm-commenting-service/pkg/http/rest/validation"
	"github.com/KompiTech/itsm-commenting-service/pkg/repository"
	"github.com/julienschmidt/httprouter"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
)

// Server is a http.Handler with dependencies
type Server struct {
	Addr                    string
	URISchema               string
	router                  *httprouter.Router
	logger                  *zap.Logger
	authService             auth.Service
	userService             usersvc.Service
	adder                   adding.Service
	lister                  listing.Service
	updater                 updating.Service
	repositoryService       repository.Service
	payloadValidator        validation.PayloadValidator
	presenter               Presenter
	ExternalLocationAddress string
}

// Config contains server configuration and dependencies
type Config struct {
	Addr                    string
	URISchema               string
	Logger                  *zap.Logger
	AuthService             auth.Service
	UserService             usersvc.Service
	AddingService           adding.Service
	ListingService          listing.Service
	UpdatingService         updating.Service
	RepositoryService       repository.Service
	PayloadValidator        validation.PayloadValidator
	ExternalLocationAddress string
}

// NewServer creates new server with the necessary dependencies
func NewServer(cfg Config) *Server {
	r := httprouter.New()

	URISchema := "http://"
	if cfg.URISchema != "" {
		URISchema = cfg.URISchema
	}

	s := &Server{
		Addr:                    cfg.Addr,
		URISchema:               URISchema,
		router:                  r,
		logger:                  cfg.Logger,
		authService:             cfg.AuthService,
		userService:             cfg.UserService,
		adder:                   cfg.AddingService,
		lister:                  cfg.ListingService,
		updater:                 cfg.UpdatingService,
		repositoryService:       cfg.RepositoryService,
		payloadValidator:        cfg.PayloadValidator,
		presenter:               NewPresenter(cfg.Logger, cfg.ExternalLocationAddress),
		ExternalLocationAddress: cfg.ExternalLocationAddress,
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

type channelIDType int

var channelIDKey channelIDType

// channelIDFromRequest returns channelID stored in request's context, if any.
func channelIDFromRequest(r *http.Request) (string, bool) {
	ch, ok := r.Context().Value(channelIDKey).(string)
	return ch, ok
}

// assertChannelID writes error message to response and returns error if channelID cannot be determined,
// otherwise it returns channelID
func (s Server) assertChannelID(w http.ResponseWriter, r *http.Request) (string, error) {
	channelID, ok := channelIDFromRequest(r)
	if !ok {
		eMsg := "could not get channel ID from context"
		s.logger.Error(eMsg)
		s.presenter.WriteError(w, eMsg, http.StatusInternalServerError)
		return "", errors.New(eMsg)
	}

	if channelID == "" {
		eMsg := "empty channel ID in context"
		s.logger.Error(eMsg)
		s.presenter.WriteError(w, "'grpc-metadata-space' header missing or invalid", http.StatusUnauthorized)
		return "", errors.New(eMsg)
	}

	return channelID, nil
}

type authType int

var authKey authType

// authTokenFromRequest returns authorization token stored in request's context, if any.
func authTokenFromRequest(r *http.Request) (string, bool) {
	ch, ok := r.Context().Value(authKey).(string)
	return ch, ok
}

// assertAuthToken writes error message to response and returns error if authorization token cannot be determined,
// otherwise it returns authorization token
func (s Server) assertAuthToken(w http.ResponseWriter, r *http.Request) (string, error) {
	authToken, ok := authTokenFromRequest(r)
	if !ok {
		eMsg := "could not get authorization token from context"
		s.logger.Error(eMsg)
		s.presenter.WriteError(w, eMsg, http.StatusInternalServerError)
		return "", errors.New(eMsg)
	}

	if authToken == "" {
		eMsg := "empty authorization token in context"
		s.logger.Error(eMsg)
		s.presenter.WriteError(w, "'authorization' header missing or invalid", http.StatusUnauthorized)
		return "", errors.New(eMsg)
	}

	return authToken, nil
}

// authorize checks if user is authorized to perform action on asset,
// otherwise it writes error message to response and returns error to notify calling handler to stop execution
func (s *Server) authorize(handlerName, assetType string, action auth.Action, w http.ResponseWriter, r *http.Request) error {
	span, ctx := opentracing.StartSpanFromContext(r.Context(), "itsm-commenting-service-authorize")
	defer span.Finish()

	r = r.WithContext(ctx)

	authToken, err := s.assertAuthToken(w, r)
	if err != nil {
		return err
	}

	channelID, err := s.assertChannelID(w, r)
	if err != nil {
		return err
	}

	if onBehalf := r.Header.Get("on_behalf"); onBehalf != "" {
		if action, err = action.OnBehalf(); err != nil {
			eMsg := fmt.Sprintf("Authorization failed: %v", err)
			s.presenter.WriteError(w, eMsg, http.StatusInternalServerError)
			return err
		}
	}

	authorized, err := s.authService.Enforce(assetType, action, channelID, authToken)
	if err != nil {
		s.logger.Error(fmt.Sprintf("%s handler failed", handlerName), zap.Error(err))
		eMsg := fmt.Sprintf("Authorization failed: %v", err)
		s.presenter.WriteError(w, eMsg, http.StatusInternalServerError)
		return err
	}

	if !authorized {
		eMsg := fmt.Sprintf("Authorization failed, action forbidden (%s, %s)", assetType, action)
		s.logger.Warn(fmt.Sprintf("%s handler failed", handlerName), zap.String("msg", eMsg))
		s.presenter.WriteError(w, eMsg, http.StatusForbidden)
		return errors.New(eMsg)
	}

	return nil
}
