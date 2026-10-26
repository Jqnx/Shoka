package api

import (
	"Shoka/internal/api/handlers"
	"Shoka/internal/api/response"
	"Shoka/internal/auth"
	"net/http"

	_ "Shoka/docs" // register the generated Swagger spec with http-swagger

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// MountMiddleware installs the global chi middleware chain: request IDs for
// log correlation, access logging, panic recovery, and permissive CORS so the
// browser frontend can call the API directly.
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

// apiHandlers bundles the handlers the router is wired to, so each
// registration group below takes one argument instead of a dozen. They are
// constructed once per router and are safe to share across requests.
type apiHandlers struct {
	test           *handlers.TestHandler
	archive        *handlers.ArchiveHandler
	metadata       *handlers.MetadataHandler
	admin          *handlers.AdminHandler
	job            *handlers.JobHandler
	library        *handlers.LibraryHandler
	tag            *handlers.TagHandler
	character      *handlers.CharacterHandler
	parody         *handlers.ParodyHandler
	artist         *handlers.ArtistHandler
	readerSettings *handlers.ReaderSettingsHandler
	search         *handlers.SearchHandler
	userAccess     *handlers.UserAccessHandler
}

// newAPIHandlers constructs one handler per route area from the server's
// shared services.
func (s *Server) newAPIHandlers() apiHandlers {
	return apiHandlers{
		test:           handlers.NewTestHandler(s.Queries, s.Processor, s.Log),
		archive:        handlers.NewArchiveHandler(s.Queries, s.DB, s.Log, s.Processor, s.Cache, s.Queue, s.Thumbnails, s.Pipeline),
		metadata:       handlers.NewMetadataHandler(s.Queries, s.DB, s.Pipeline, s.Processor, s.Log),
		admin:          handlers.NewAdminHandler(s.Queries, s.Maintenance, s.Log),
		job:            handlers.NewJobHandler(s.Queries, s.Log),
		library:        handlers.NewLibraryHandler(s.Queries, s.DB, s.Queue, s.Libraries, s.Processor, s.Pipeline, s.Log),
		tag:            handlers.NewTagHandler(s.Queries, s.Log, s.Processor),
		character:      handlers.NewCharacterHandler(s.Queries, s.Log, s.Processor),
		parody:         handlers.NewParodyHandler(s.Queries, s.Log, s.Processor),
		artist:         handlers.NewArtistHandler(s.Queries, s.DB, s.Log, s.Processor),
		readerSettings: handlers.NewReaderSettingsHandler(s.Queries, s.Log),
		search:         handlers.NewSearchHandler(s.Queries, s.DB, s.Processor, s.Log),
		userAccess:     handlers.NewUserAccessHandler(s.Queries, s.DB, s.Log),
	}
}

// MountHandlers registers every API route. Everything except the docs and
// health endpoints sits in a group behind authMiddleware, and individual
// routes add their own requirePermission gate.
//
// Registration is split by area below, and the order the areas are mounted in
// is the order the routes are registered in - which is part of the contract,
// see the note on /api/archives/bulk.
func (s *Server) MountHandlers() {
	s.mountPublicRoutes()

	s.Router.Group(func(r chi.Router) {
		r.Use(authMiddleware(s.DB))

		h := s.newAPIHandlers()

		mountBrowseRoutes(r, h)
		mountEntityRoutes(r, h)
		mountArchiveDetailRoutes(r, h)
		mountAdminRoutes(r, h)
		mountTestRoutes(r, h)
	})
}

// mountPublicRoutes registers the endpoints that need no authentication.
func (s *Server) mountPublicRoutes() {
	s.Router.Get("/docs/*", httpSwagger.Handler(
		httpSwagger.URL("/docs/doc.json"),
	))

	s.Router.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
}

// mountBrowseRoutes registers the read-only browse surface: metadata sources,
// libraries, cross-category search, and the archive list with its bulk
// sub-routes.
func mountBrowseRoutes(r chi.Router, h apiHandlers) {
	r.Get("/api/metadata/sources", h.metadata.GetSources)
	r.Get("/api/libraries", h.library.GetLibraries)
	r.Get("/api/libraries/types", h.library.GetLibraryTypes)
	r.Get("/api/search", h.search.Search)
	r.Get("/api/archives", h.archive.GetArchives)
	r.Get("/api/archives/sort-options", h.archive.GetArchiveSortOptions)
	r.Get("/api/archives/categories", h.archive.GetCategories)
	r.Get("/api/archives/languages", h.archive.GetLanguages)
	r.Get("/api/archives/recently-read", h.archive.GetRecentlyRead)
	r.Get("/api/archives/favorites", h.archive.GetFavorites)
	r.Get("/api/archives/favorites/pages", h.archive.GetFavoritePages)

	// Registered before the /api/archives/{id} routes below: chi matches
	// static segments ahead of wildcards, so "bulk" can't be swallowed as
	// an archive id, but keeping them adjacent makes that ordering
	// obvious to the next reader.
	r.Route("/api/archives/bulk", func(r chi.Router) {
		r.With(requirePermission(auth.ArchiveUpdate)).Patch("/", h.archive.BulkUpdateArchives)
		r.Put("/progress", h.archive.BulkUpdateProgress)
		r.With(requirePermission(auth.MetadataFetch)).Post("/metadata", h.archive.BulkFetchMetadata)
	})
}

// mountEntityRoutes registers the taxonomy endpoints - tags, characters,
// parodies and artists - plus the reader settings endpoints, in the order the
// two static-before-wildcard families within each require (all vs {id} vs
// {name}).
func mountEntityRoutes(r chi.Router, h apiHandlers) {
	r.Get("/api/tags", h.tag.GetTags)
	r.Get("/api/tags/all", h.tag.GetAllTags)
	r.Route("/api/tags/{id}", func(r chi.Router) {
		r.With(requirePermission(auth.TaxonomyUpdate)).Patch("/", h.tag.UpdateTagDescription)
		r.With(requirePermission(auth.TaxonomyDelete)).Delete("/", h.tag.DeleteTag)
	})
	r.Get("/api/tags/{name}", h.tag.GetArchivesByTag)

	r.Get("/api/characters", h.character.GetCharacters)
	r.Get("/api/characters/all", h.character.GetAllCharacters)
	r.With(requirePermission(auth.TaxonomyDelete)).Delete("/api/characters/{id}", h.character.DeleteCharacter)
	r.Get("/api/characters/{name}", h.character.GetArchivesByCharacter)

	r.Get("/api/parodies", h.parody.GetParodies)
	r.Get("/api/parodies/all", h.parody.GetAllParodies)
	r.With(requirePermission(auth.TaxonomyDelete)).Delete("/api/parodies/{id}", h.parody.DeleteParody)
	r.Get("/api/parodies/{name}", h.parody.GetArchivesByParody)

	r.Route("/api/reader-settings", func(r chi.Router) {
		r.Get("/", h.readerSettings.GetReaderSettings)
		r.Patch("/", h.readerSettings.UpdateReaderSettings)
	})

	r.Get("/api/artists", h.artist.GetArtists)
	r.With(requirePermission(auth.TaxonomyCreate)).Post("/api/artists", h.artist.CreateArtist)
	r.Get("/api/artists/all", h.artist.GetAllArtists)
	r.Route("/api/artists/{id}", func(r chi.Router) {
		r.Get("/", h.artist.GetArtist)
		r.With(requirePermission(auth.TaxonomyUpdate)).Patch("/", h.artist.UpdateArtist)
		r.With(requirePermission(auth.TaxonomyDelete)).Delete("/", h.artist.DeleteArtist)
		r.Get("/archives", h.artist.GetArchivesByArtist)
	})
}

// mountArchiveDetailRoutes registers the single-archive routes and their
// metadata sub-resources. Mounted after mountBrowseRoutes so the static
// /api/archives/* routes it shares a prefix with are matched first.
func mountArchiveDetailRoutes(r chi.Router, h apiHandlers) {
	r.Route("/api/archives/{id}", func(r chi.Router) {
		r.Get("/", h.archive.GetArchive)
		r.With(requirePermission(auth.ArchiveUpdate)).Patch("/", h.archive.UpdateArchive)
		r.With(requirePermission(auth.ArchiveDelete)).Delete("/", h.archive.DeleteArchive)
		r.Get("/cover", h.archive.GetCover)
		r.With(requirePermission(auth.ArchiveCover)).Put("/cover", h.archive.SetCover)
		r.Get("/pages/{index}", h.archive.GetPage)
		r.Get("/pages/{index}/thumbnail", h.archive.GetPageThumbnail)
		r.With(requirePermission(auth.ArchiveThumbnail)).Post("/thumbnails", h.archive.GenerateThumbnails)
		r.Get("/thumbnails/events", h.archive.StreamThumbnailEvents)
		r.Put("/progress", h.archive.UpdateProgress)
		r.Delete("/progress", h.archive.DeleteProgress)
		r.Put("/favorite", h.archive.AddFavorite)
		r.Delete("/favorite", h.archive.RemoveFavorite)
		r.Get("/favorite-pages", h.archive.GetFavoritePageIndices)
		r.Put("/pages/{index}/favorite", h.archive.AddFavoritePage)
		r.Delete("/pages/{index}/favorite", h.archive.RemoveFavoritePage)
		r.Put("/rating", h.archive.SetRating)
		r.Delete("/rating", h.archive.RemoveRating)
	})
	r.Route("/api/archives/{id}/metadata", func(r chi.Router) {
		r.With(requirePermission(auth.MetadataFetch)).Post("/", h.metadata.FetchMetadata)
		r.With(requirePermission(auth.MetadataWrite)).Post("/save", h.metadata.SaveMetadataToFile)
		r.With(requirePermission(auth.MetadataFetch)).Post("/{source}", h.metadata.FetchMetadataFromSource)
		r.With(requirePermission(auth.MetadataFetch)).Get("/{source}/search", h.metadata.SearchMetadataSource)
		r.With(requirePermission(auth.MetadataApply)).
			Post("/{source}/{source_id}", h.metadata.ApplyMetadataFromSource)
	})
}

// mountAdminRoutes registers the cross-library maintenance endpoints: covers,
// perceptual hashes, duplicate detection, the job queue, and library
// administration.
func mountAdminRoutes(r chi.Router, h apiHandlers) {
	r.Route("/api/admin", func(r chi.Router) {
		r.With(requirePermission(auth.MaintenanceRun)).Post("/covers", h.admin.GenerateCovers)
		r.With(requirePermission(auth.MaintenanceRun)).Post("/phashes", h.admin.GeneratePHashes)
		r.With(requirePermission(auth.MaintenanceRead)).Get("/maintenance", h.admin.GetMaintenanceTasks)
		r.With(requirePermission(auth.MaintenanceRead)).Get("/duplicates", h.admin.GetDuplicates)
		r.With(requirePermission(auth.JobRead)).Get("/jobs", h.job.GetJobs)
		r.With(requirePermission(auth.JobRead)).Get("/jobs/events", h.job.StreamJobEvents)

		r.Route("/users/{id}", func(r chi.Router) {
			r.With(requirePermission(auth.UserGet)).Get("/library-access", h.userAccess.GetLibraryAccess)
			r.With(requirePermission(auth.UserUpdate)).Put("/library-access", h.userAccess.SetLibraryAccess)
		})

		r.Route("/libraries", func(r chi.Router) {
			r.With(requirePermission(auth.LibraryCreate)).Post("/", h.library.CreateLibrary)
			r.With(requirePermission(auth.LibraryCreate)).Get("/browse", h.library.BrowseLibraryPath)
			r.With(requirePermission(auth.LibraryScan)).Post("/scan", h.library.ScanAllLibraries)
			r.Route("/{id}", func(r chi.Router) {
				r.With(requirePermission(auth.LibraryUpdate)).Patch("/", h.library.UpdateLibrary)
				r.With(requirePermission(auth.LibraryDelete)).Delete("/", h.library.DeleteLibrary)
				r.With(requirePermission(auth.LibraryScan)).Post("/scan", h.library.ScanLibrary)
				r.With(requirePermission(auth.MaintenanceRun)).Post("/covers", h.library.GenerateLibraryCovers)
				r.With(requirePermission(auth.LibraryRead)).Get("/sources", h.library.GetLibrarySources)
				r.With(requirePermission(auth.LibraryUpdate)).
					Put("/sources/order", h.library.ReorderLibrarySources)
				r.With(requirePermission(auth.LibraryUpdate)).
					Patch("/sources/{source}", h.library.UpdateLibrarySource)
			})
		})
	})
}

// mountTestRoutes registers the fixture endpoints used by the test suite.
func mountTestRoutes(r chi.Router, h apiHandlers) {
	r.Route("/api/test", func(r chi.Router) {
		r.Get("/", h.test.GetTest)
	})
}
