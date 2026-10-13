package api

import (
	"Shoka/internal/database/sqlc"
	"Shoka/internal/events"
	"Shoka/internal/image"
	"Shoka/internal/jobs"
	"Shoka/internal/library"
	"Shoka/internal/metadata"
	"database/sql"
	"log/slog"

	"github.com/go-chi/chi/v5"
)

// Server owns the chi router together with the shared services every handler
// borrows: the SQL queries and DB pool, job queue, image cache and processor,
// metadata pipeline, library manager, thumbnail broadcaster, and logger.
type Server struct {
	Router     *chi.Mux
	Queries    *sqlc.Queries
	Log        *slog.Logger
	Queue      *jobs.Queue
	Cache      *image.Cache
	Processor  *image.Processor
	Pipeline   *metadata.Pipeline
	Libraries  *library.Manager
	Thumbnails *events.ThumbnailBroadcaster
	DB         *sql.DB
}

// New assembles a Server from the already-constructed services and mounts the
// middleware chain and all routes before returning it.
func New(queries *sqlc.Queries, logger *slog.Logger, queue *jobs.Queue, cache *image.Cache, processor *image.Processor, pipeline *metadata.Pipeline, libraries *library.Manager, thumbnails *events.ThumbnailBroadcaster, db *sql.DB) *Server {
	router := chi.NewRouter()
	server := &Server{
		Router:     router,
		Queries:    queries,
		Log:        logger.With("component", "api"),
		Queue:      queue,
		Cache:      cache,
		Processor:  processor,
		Pipeline:   pipeline,
		Libraries:  libraries,
		Thumbnails: thumbnails,
		DB:         db,
	}
	server.MountMiddleware()
	server.MountHandlers()

	return server
}
