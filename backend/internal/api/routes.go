package api

import (
	"Shoka/internal/api/handlers"
	"Shoka/internal/api/response"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (s *Server) MountMiddleware() {
	s.Router.Use(middleware.RequestID)
	s.Router.Use(middleware.Logger)
	s.Router.Use(middleware.Recoverer)
	// s.Router.Use(httprate.LimitByIP(30, time.Minute))
}

func (s *Server) MountHandlers() {
	s.Router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	s.Router.Group(func(r chi.Router) {
		testHandler := handlers.NewTestHandler(s.Queries, s.Processor, s.Log)

		r.Route("/api/test", func(r chi.Router) {
			r.Get("/", testHandler.GetTest)
		})
	})
}
