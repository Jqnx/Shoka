package server

import (
	"Shoka/internal/config"
	"Shoka/internal/models"
	"Shoka/internal/workers"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

func (s *Server) scanMetadataHandler(c *gin.Context) {
	ctx := context.Background()
	id := c.Param("id")
	src := c.Query("source")

	archive, err := s.repo.GetArchiveByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, &models.Response{
				Status:  "error",
				Message: config.ErrArchiveNotFound.Error(),
			})
			return
		} else {
			c.JSON(http.StatusInternalServerError, &models.Response{
				Status:  "error",
				Message: err.Error(),
			})
			return
		}
	}

	w := workers.NewWorkers(s.app, false, ctx)
	go w.Metadata(&archive, src)
	c.JSON(http.StatusOK, "success")
}
