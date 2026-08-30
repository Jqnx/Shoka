package api

import (
	"Shoka/internal/api/handlers"
	"Shoka/internal/api/response"
	"net/http"

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
		archiveHandler := handlers.NewArchiveHandler(s.Queries, s.DB, s.Log, s.Processor, s.Cache, s.Queue, s.Thumbnails, s.Pipeline)
		metadataHandler := handlers.NewMetadataHandler(s.Queries, s.DB, s.Pipeline, s.Processor, s.Log)
		adminHandler := handlers.NewAdminHandler(s.Queries, s.Queue, s.Log)
		jobHandler := handlers.NewJobHandler(s.Queries, s.Log)
		libraryHandler := handlers.NewLibraryHandler(s.Queries, s.Queue, s.Libraries, s.Log)
		tagHandler := handlers.NewTagHandler(s.Queries, s.Log, s.Processor)
		characterHandler := handlers.NewCharacterHandler(s.Queries, s.Log, s.Processor)
		parodyHandler := handlers.NewParodyHandler(s.Queries, s.Log, s.Processor)
		artistHandler := handlers.NewArtistHandler(s.Queries, s.DB, s.Log, s.Processor)
		readerSettingsHandler := handlers.NewReaderSettingsHandler(s.Queries, s.Log)
		searchHandler := handlers.NewSearchHandler(s.Queries, s.DB, s.Processor, s.Log)

		r.Get("/api/metadata/sources", metadataHandler.GetSources)
		r.Get("/api/libraries", libraryHandler.GetLibraries)
		r.Get("/api/libraries/types", libraryHandler.GetLibraryTypes)
		r.Get("/api/search", searchHandler.Search)
		r.Get("/api/archives", archiveHandler.GetArchives)
		r.Get("/api/archives/sort-options", archiveHandler.GetArchiveSortOptions)
		r.Get("/api/archives/categories", archiveHandler.GetCategories)
		r.Get("/api/archives/languages", archiveHandler.GetLanguages)
		r.Get("/api/archives/recently-read", archiveHandler.GetRecentlyRead)
		r.Get("/api/archives/favorites", archiveHandler.GetFavorites)

		// Registered before the /api/archives/{id} routes below: chi matches
		// static segments ahead of wildcards, so "bulk" can't be swallowed as
		// an archive id, but keeping them adjacent makes that ordering
		// obvious to the next reader.
		r.Route("/api/archives/bulk", func(r chi.Router) {
			r.Patch("/", archiveHandler.BulkUpdateArchives)
			r.Put("/progress", archiveHandler.BulkUpdateProgress)
			r.Post("/metadata", archiveHandler.BulkFetchMetadata)
		})
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

		r.Route("/api/reader-settings", func(r chi.Router) {
			r.Get("/", readerSettingsHandler.GetReaderSettings)
			r.Patch("/", readerSettingsHandler.UpdateReaderSettings)
		})

		r.Get("/api/artists", artistHandler.GetArtists)
		r.Post("/api/artists", artistHandler.CreateArtist)
		r.Get("/api/artists/all", artistHandler.GetAllArtists)
		r.Route("/api/artists/{id}", func(r chi.Router) {
			r.Get("/", artistHandler.GetArtist)
			r.Patch("/", artistHandler.UpdateArtist)
			r.Delete("/", artistHandler.DeleteArtist)
			r.Get("/archives", artistHandler.GetArchivesByArtist)
		})
		r.Route("/api/archives/{id}", func(r chi.Router) {
			r.Get("/", archiveHandler.GetArchive)
			r.Patch("/", archiveHandler.UpdateArchive)
			r.Delete("/", archiveHandler.DeleteArchive)
			r.Get("/cover", archiveHandler.GetCover)
			r.Get("/pages/{index}", archiveHandler.GetPage)
			r.Get("/pages/{index}/thumbnail", archiveHandler.GetPageThumbnail)
			r.Post("/thumbnails", archiveHandler.GenerateThumbnails)
			r.Get("/thumbnails/events", archiveHandler.StreamThumbnailEvents)
			r.Put("/progress", archiveHandler.UpdateProgress)
			r.Delete("/progress", archiveHandler.DeleteProgress)
			r.Put("/favorite", archiveHandler.AddFavorite)
			r.Delete("/favorite", archiveHandler.RemoveFavorite)
			r.Put("/rating", archiveHandler.SetRating)
			r.Delete("/rating", archiveHandler.RemoveRating)
		})
		r.Route("/api/archives/{id}/metadata", func(r chi.Router) {
			r.Post("/", metadataHandler.FetchMetadata)
			r.Post("/save", metadataHandler.SaveMetadataToFile)
			r.Post("/{source}", metadataHandler.FetchMetadataFromSource)
			r.Get("/{source}/search", metadataHandler.SearchMetadataSource)
			r.Post("/{source}/{source_id}", metadataHandler.ApplyMetadataFromSource)
		})

		r.Route("/api/admin", func(r chi.Router) {
			r.Post("/covers", adminHandler.GenerateCovers)
			r.Post("/phashes", adminHandler.GeneratePHashes)
			r.Get("/duplicates", adminHandler.GetDuplicates)
			r.Get("/jobs", jobHandler.GetJobs)
			r.Get("/jobs/events", jobHandler.StreamJobEvents)

			r.Route("/libraries", func(r chi.Router) {
				r.Post("/", libraryHandler.CreateLibrary)
				r.Route("/{id}", func(r chi.Router) {
					r.Patch("/", libraryHandler.UpdateLibrary)
					r.Delete("/", libraryHandler.DeleteLibrary)
					r.Post("/scan", libraryHandler.ScanLibrary)
					r.Post("/covers", libraryHandler.GenerateLibraryCovers)
					r.Get("/sources", libraryHandler.GetLibrarySources)
					r.Patch("/sources/{source}", libraryHandler.UpdateLibrarySource)
				})
			})
		})

		r.Route("/api/test", func(r chi.Router) {
			r.Get("/", testHandler.GetTest)
		})
	})
}
