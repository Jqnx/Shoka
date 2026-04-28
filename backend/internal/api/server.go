package api

import (
	"Shoka/internal/database/sqlc"
	"Shoka/internal/image"
	"Shoka/internal/jobs"
	"Shoka/internal/metadata"
	"log/slog"

	"github.com/go-chi/chi/v5"
)

type Server struct {
	Router    *chi.Mux
	Queries   *sqlc.Queries
	Log       *slog.Logger
	Queue     *jobs.Queue
	Cache     *image.Cache
	Processor *image.Processor
	Pipeline  *metadata.Pipeline
}

func New(queries *sqlc.Queries, logger *slog.Logger, queue *jobs.Queue, cache *image.Cache, processor *image.Processor, pipeline *metadata.Pipeline) *Server {
	router := chi.NewRouter()
	server := &Server{
		Router:    router,
		Queries:   queries,
		Log:       logger.With("component", "api"),
		Queue:     queue,
		Cache:     cache,
		Processor: processor,
		Pipeline:  pipeline,
	}
	server.MountMiddleware()
	server.MountHandlers()

	return server
}
