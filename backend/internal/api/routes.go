package api

import (
	"net/http"

	"Shoka/internal/api/handlers"
	"Shoka/internal/api/response"

	_ "Shoka/docs"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func (s *Server) MountMiddleware() {
	s.Router.Use(middleware.RequestID)
	s.Router.Use(middleware.Logger)
	s.Router.Use(middleware.Recoverer)
	// s.Router.Use(httprate.LimitByIP(30, time.Minute))
}

func (s *Server) MountHandlers() {
	s.Router.Get("/docs/*", httpSwagger.Handler(
		httpSwagger.URL("/docs/doc.json"),
	))

	s.Router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	s.Router.Group(func(r chi.Router) {
		r.Use(authMiddleware(s.DB))

		testHandler := handlers.NewTestHandler(s.Queries, s.Processor, s.Log)
		archiveHandler := handlers.NewArchiveHandler(s.Queries, s.Log, s.Processor)
		metadataHandler := handlers.NewMetadataHandler(s.Queries, s.DB, s.Pipeline, s.Log)
		adminHandler := handlers.NewAdminHandler(s.Queries, s.Log)

		r.Get("/api/metadata/sources", metadataHandler.GetSources)
		r.Get("/api/archives", archiveHandler.GetArchives)
		r.Route("/api/archives/{id}", func(r chi.Router) {
			r.Get("/", archiveHandler.GetArchive)
		})
		r.Route("/api/archives/{id}/metadata", func(r chi.Router) {
			r.Post("/", metadataHandler.FetchMetadata)
			r.Post("/save", metadataHandler.SaveMetadataToFile)
			r.Post("/{source}", metadataHandler.FetchMetadataFromSource)
			r.Get("/{source}/search", metadataHandler.SearchMetadataSource)
			r.Post("/{source}/{source_id}", metadataHandler.ApplyMetadataFromSource)
		})

		r.Route("/api/admin", func(r chi.Router) {
			r.Patch("/sources/{source}", adminHandler.UpdateSourceEnabled)
		})

		r.Route("/api/test", func(r chi.Router) {
			r.Get("/", testHandler.GetTest)
		})
	})
}
