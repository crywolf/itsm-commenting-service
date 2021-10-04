package rest

import (
	"embed"
	"net/http"

	"github.com/KompiTech/go-toolkit/common"
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment"
	"github.com/go-openapi/runtime/middleware"
	"github.com/justinas/alice"
	"github.com/opentracing/opentracing-go"
)

//go:embed swagger.yaml
var swaggerFS embed.FS

func (s *Server) routes() {
	router := s.router

	// tracing handler
	var (
		traceMW = common.Trace{
			Tracer: opentracing.GlobalTracer(),
		}
		rIDMW common.RequestID
	)

	chain := alice.New(rIDMW.RequestIDMiddleware, traceMW.TraceMiddleware)

	chain.Then(router)

	// comments
	router.GET("/comments/:id", s.GetComment(comment.AssetTypeComment))
	router.GET("/comments", s.QueryComments(comment.AssetTypeComment))

	router.POST("/comments", s.AddUserInfo(s.AddComment(comment.AssetTypeComment), s.userService))
	router.POST("/comments/:id/read_by", s.AddUserInfo(s.MarkCommentAsReadBy(comment.AssetTypeComment), s.userService))

	// worknotes
	router.GET("/worknotes/:id", s.GetComment(comment.AssetTypeWorknote))
	router.GET("/worknotes", s.QueryComments(comment.AssetTypeWorknote))

	router.POST("/worknotes", s.AddUserInfo(s.AddComment(comment.AssetTypeWorknote), s.userService))
	router.POST("/worknotes/:id/read_by", s.AddUserInfo(s.MarkCommentAsReadBy(comment.AssetTypeWorknote), s.userService))

	// databases creation
	router.POST("/databases", s.CreateDatabases())

	// API documentation
	opts := middleware.RedocOpts{Path: "/docs", SpecURL: "/swagger.yaml", Title: "Commenting service API documentation"}
	docsHandler := middleware.Redoc(opts, nil)
	// handlers for API documentation
	router.Handler(http.MethodGet, "/docs", docsHandler)
	router.Handler(http.MethodGet, "/swagger.yaml", http.FileServer(http.FS(swaggerFS)))

	// default Not Found handler
	router.NotFound = http.HandlerFunc(s.JSONNotFoundError)
}

// JSONNotFoundError replies to the request with the 404 page not found general error message
// in JSON format and sets correct header and HTTP code
func (s Server) JSONNotFoundError(w http.ResponseWriter, _ *http.Request) {
	s.presenter.WriteError(w, "404 page not found", http.StatusNotFound)
}
