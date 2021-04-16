package rest

import (
	"fmt"
	"net/http"

	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment/adding"
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment/listing"
	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
)

// Server is a http.Handler with dependencies
type Server struct {
	Addr   string
	router *httprouter.Router
	logger *zap.Logger

	adder  adding.Service
	lister listing.Service
}

// Config contains server configuration and dependencies
type Config struct {
	Addr           string
	Logger         *zap.Logger
	AddingService  adding.Service
	ListingService listing.Service
}

// NewServer creates new server with the necessary dependencies
func NewServer(cfg Config) *Server {
	r := httprouter.New()

	s := &Server{
		Addr:   cfg.Addr,
		router: r,
		logger: cfg.Logger,
		adder:  cfg.AddingService,
		lister: cfg.ListingService}
	s.routes()

	return s
}

// ServeHTTP makes the server implement the http.Handler interface
func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
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
