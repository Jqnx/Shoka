package api

import (
	"database/sql"
	"log/slog"

	"Shoka/internal/database/sqlc"
	"Shoka/internal/events"
	"Shoka/internal/image"
	"Shoka/internal/jobs"
	"Shoka/internal/library"
	"Shoka/internal/metadata"

	"github.com/go-chi/chi/v5"
)

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
