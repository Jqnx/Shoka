package api

import (
	"net/http"

	"Shoka/internal/api/handlers"
	"Shoka/internal/api/response"

	_ "Shoka/docs"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func (s *Server) MountMiddleware() {
	s.Router.Use(middleware.RequestID)
	s.Router.Use(middleware.Logger)
	s.Router.Use(middleware.Recoverer)
	s.Router.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		MaxAge:         300,
	}))
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
		archiveHandler := handlers.NewArchiveHandler(s.Queries, s.DB, s.Log, s.Processor, s.Cache)
		metadataHandler := handlers.NewMetadataHandler(s.Queries, s.DB, s.Pipeline, s.Log)
		adminHandler := handlers.NewAdminHandler(s.Queries, s.Queue, s.Log)
		tagHandler := handlers.NewTagHandler(s.Queries, s.Log, s.Processor)
		characterHandler := handlers.NewCharacterHandler(s.Queries, s.Log, s.Processor)
		parodyHandler := handlers.NewParodyHandler(s.Queries, s.Log, s.Processor)

		r.Get("/api/metadata/sources", metadataHandler.GetSources)
		r.Get("/api/archives", archiveHandler.GetArchives)
		r.Get("/api/tags", tagHandler.GetTags)
		r.Get("/api/tags/all", tagHandler.GetAllTags)
		r.Route("/api/tags/{id}", func(r chi.Router) {
			r.Patch("/", tagHandler.UpdateTagDescription)
			r.Delete("/", tagHandler.DeleteTag)
		})
		r.Get("/api/tags/{name}", tagHandler.GetArchivesByTag)

		r.Get("/api/characters", characterHandler.GetCharacters)
		r.Get("/api/characters/all", characterHandler.GetAllCharacters)
		r.Delete("/api/characters/{id}", characterHandler.DeleteCharacter)
		r.Get("/api/characters/{name}", characterHandler.GetArchivesByCharacter)

		r.Get("/api/parodies", parodyHandler.GetParodies)
		r.Get("/api/parodies/all", parodyHandler.GetAllParodies)
		r.Delete("/api/parodies/{id}", parodyHandler.DeleteParody)
		r.Get("/api/parodies/{name}", parodyHandler.GetArchivesByParody)
		r.Route("/api/archives/{id}", func(r chi.Router) {
			r.Get("/", archiveHandler.GetArchive)
			r.Get("/cover", archiveHandler.GetCover)
			r.Get("/pages/{index}", archiveHandler.GetPage)
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
			r.Post("/covers", adminHandler.GenerateCovers)
		})

		r.Route("/api/test", func(r chi.Router) {
			r.Get("/", testHandler.GetTest)
		})
	})
}
