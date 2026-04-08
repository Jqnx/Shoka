package api

import (
	"Shoka/internal/db/sqlc"

	"github.com/go-chi/chi/v5"
)

type Server struct {
	Router  *chi.Mux
	Queries *sqlc.Queries
}

func New(queries *sqlc.Queries) *Server {
	router := chi.NewRouter()
	return &Server{
		Router:  router,
		Queries: queries,
	}
}
