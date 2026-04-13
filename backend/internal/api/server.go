package api

import (
	"log/slog"

	"Shoka/internal/database/sqlc"
	"Shoka/internal/image"
	"Shoka/internal/jobs"

	"github.com/go-chi/chi/v5"
)

type Server struct {
	Router    *chi.Mux
	Queries   *sqlc.Queries
	Log       *slog.Logger
	Queue     *jobs.Queue
	Cache     *image.Cache
	Processor *image.Processor
}

func New(queries *sqlc.Queries, logger *slog.Logger, queue *jobs.Queue, cache *image.Cache, processor *image.Processor) *Server {
	router := chi.NewRouter()
	server := &Server{
		Router:    router,
		Queries:   queries,
		Log:       logger.With("component", "api"),
		Queue:     queue,
		Cache:     cache,
		Processor: processor,
	}
	server.MountMiddleware()
	server.MountHandlers()

	return server
}
