package rest

import "net/http"

func (s *Server) routes() {
	router := s.router

	router.GET("/comments/:id", s.GetComment())
	router.GET("/comments", s.QueryComments())
	router.POST("/comments", s.AddComment())

	//	router.POST("/comments/query", s.QueryComments())

	//router.GET("/worknotes/:id", s.GetWorknote())
	//router.POST("/worknotes", s.AddWorknote())

	router.NotFound = http.HandlerFunc(s.JSONNotFoundError)
}

// JSONNotFoundError replies to the request with the 404 page not found general error message
// in JSON format and sets correct header and HTTP code
func (s Server) JSONNotFoundError(w http.ResponseWriter, _ *http.Request) {
	s.JSONError(w, "404 page not found", http.StatusNotFound)
}
