package handlers

import (
	"Shoka/internal/api/response"
	"Shoka/internal/database/sqlc"
	"Shoka/internal/image"
	"log/slog"
	"net/http"
)

// TestHandler exposes a trivial endpoint used to verify the router and
// response helpers are wired up correctly.
type TestHandler struct {
	queries   *sqlc.Queries
	processor *image.Processor
	log       *slog.Logger
}

// NewTestHandler returns a TestHandler backed by the given queries, image
// processor, and logger.
func NewTestHandler(queries *sqlc.Queries, processor *image.Processor, log *slog.Logger) *TestHandler {
	return &TestHandler{
		queries:   queries,
		processor: processor,
		log:       log.With("handler", "test"),
	}
}

// GetTest godoc
//
// @Summary Test
// @Description Example Description
// @Tags test
// @Produce json
// @Param example query string false "example"
// @Success 200 {object} map[string]string
// @Failure 400 {object} response.Error
// @Router /api/test [get].
func (t *TestHandler) GetTest(w http.ResponseWriter, _ *http.Request) {
	response.JSON(w, http.StatusOK, map[string]string{"status": "test"})
}
