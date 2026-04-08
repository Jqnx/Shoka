package api

import "github.com/go-chi/chi/v5/middleware"

func (s *Server) MountHandlers() {
	s.Router.Use(middleware.Logger)
}
