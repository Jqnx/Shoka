package server

import (
	"Shoka/internal/models"
	"Shoka/internal/workers"
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
)

func (s *Server) generateCoverHandler(c *gin.Context) {
	id := c.Param("id")
	force := c.Query("force")

	ctx := context.Background()

	arch, err := s.repo.GetArchiveByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.Response{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	ch := make(chan *asynq.TaskInfo)

	if force == "true" {
		w := workers.NewWorkers(s.app, true, ctx)
		go w.Covers(ch, &arch)
		cover := <-ch

		// Sets Location header to URI of job status
		c.Header("Location", fmt.Sprintf("/api/status/%s", cover.ID))
	} else {
		w := workers.NewWorkers(s.app, false, ctx)
		go w.Covers(ch, &arch)
		cover := <-ch
		c.Header("Location", fmt.Sprintf("/api/status/%s", cover.ID))
	}
}

func (s *Server) getCoverHandler(c *gin.Context) {
	id := c.Param("id")

	ctx := context.Background()

	arch, err := s.repo.GetArchiveByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.Response{
			Status:  "error",
			Message: err,
		})
		return
	}

	if arch.ThumbsPath == nil {
		c.JSON(http.StatusInternalServerError, &models.Response{
			Status:  "error",
			Message: "no cover found",
		})
		return
	}

	if *arch.ThumbsPath == "" {
		c.JSON(http.StatusInternalServerError, &models.Response{
			Status:  "error",
			Message: "no cover found",
		})
		return
	}

	c.File(*arch.CoverPath)
}
