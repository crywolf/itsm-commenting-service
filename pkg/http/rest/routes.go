package rest

import (
	"net/http"
)

const (
	assetTypeComment  = "comment"
	assetTypeWorknote = "worknote"
)

func (s *Server) routes() {
	router := s.router

	// comments
	router.GET("/comments/:id", s.GetComment(assetTypeComment))
	router.GET("/comments", s.QueryComments(assetTypeComment))

	router.POST("/comments", s.AddUserInfo(s.AddComment(assetTypeComment), s.userService))
	router.POST("/comments/:id/read_by", s.AddUserInfo(s.MarkAsReadBy(assetTypeComment), s.userService))

	// worknotes
	router.GET("/worknotes/:id", s.GetComment(assetTypeWorknote))
	router.GET("/worknotes", s.QueryComments(assetTypeWorknote))

	router.POST("/worknotes", s.AddUserInfo(s.AddComment(assetTypeWorknote), s.userService))
	router.POST("/worknotes/:id/read_by", s.AddUserInfo(s.MarkAsReadBy(assetTypeWorknote), s.userService))

	// databases creation
	router.POST("/databases", s.CreateDatabases())

	// default Not Found handler
	router.NotFound = http.HandlerFunc(s.JSONNotFoundError)
}

// JSONNotFoundError replies to the request with the 404 page not found general error message
// in JSON format and sets correct header and HTTP code
func (s Server) JSONNotFoundError(w http.ResponseWriter, _ *http.Request) {
	s.JSONError(w, "404 page not found", http.StatusNotFound)
}
