package rest

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment/adding"
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment/listing"
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment/updating"
	"github.com/KompiTech/itsm-commenting-service/pkg/repository"
	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
)

// Server is a http.Handler with dependencies
type Server struct {
	Addr   string
	router *httprouter.Router
	logger *zap.Logger

	userService UserService
	adder       adding.Service
	lister      listing.Service
	updater     updating.Service
	repositoryService repository.Service
}

// Config contains server configuration and dependencies
type Config struct {
	Addr            string
	Logger          *zap.Logger
	UserService     UserService
	AddingService   adding.Service
	ListingService  listing.Service
	UpdatingService updating.Service
	RepositoryService repository.Service
}

// NewServer creates new server with the necessary dependencies
func NewServer(cfg Config) *Server {
	r := httprouter.New()

	s := &Server{
		Addr:        cfg.Addr,
		router:      r,
		logger:      cfg.Logger,
		userService: cfg.UserService,
		adder:       cfg.AddingService,
		lister:      cfg.ListingService,
		updater:     cfg.UpdatingService,
		repositoryService: cfg.RepositoryService,
	}
	s.routes()

	return s
}

type channelIDType int

var channelIDKey channelIDType

// channelIDFromContext returns channelIDe stored in ctx, if any.
func channelIDFromContext(ctx context.Context) (string, bool) {
	ch, ok := ctx.Value(channelIDKey).(string)
	return ch, ok
}

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

// ServeHTTP makes the server implement the http.Handler interface
func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	channelID := r.Header.Get("grpc-metadata-space")
	ctx := context.WithValue(r.Context(), channelIDKey, channelID)

	s.router.ServeHTTP(w, r.WithContext(ctx))
}

// JSONError replies to the request with the specified error message and HTTP code.
// It encode error string as JSON object {"error":"error_string"} and sets correct header.
// It does not otherwise end the request; the caller should ensure no further  writes are done to w.
// The error message should be plain text.
func (s Server) JSONError(w http.ResponseWriter, error string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, _ = fmt.Fprintln(w, `{"error":"`+error+`"}`)
}
