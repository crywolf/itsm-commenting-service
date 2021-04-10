package rest

func (s *Server) routes() {
	router := s.router

	router.GET("/comments/:id", s.GetComment())
	router.POST("/comments", s.AddComment())

	//router.GET("/worknotes/:id", s.GetWorknote())
	//router.POST("/worknotes", s.AddWorknote())
}
