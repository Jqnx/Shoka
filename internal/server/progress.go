package server

import (
	"Shoka/internal/models"
	"Shoka/internal/repository"
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type ReadingProgress struct {
	ReadingState string
	Progress     int64
	LastRead     *time.Time
}

func GetReadingProgress(page, max int64) *ReadingProgress {
	now := time.Now()
	switch {
	case page == 0:
		return &ReadingProgress{
			ReadingState: "unread",
			Progress:     0,
			LastRead:     nil,
		}
	case page > 0 && page < max:
		return &ReadingProgress{
			ReadingState: "reading",
			Progress:     page,
			LastRead:     &now,
		}
	case page == max:
		return &ReadingProgress{
			ReadingState: "finished",
			Progress:     page,
			LastRead:     &now,
		}
	}
	return nil
}

func (s *Server) updateReadingProgressHandler(c *gin.Context) {
	ctx := context.Background()

	archive_id := c.Param("id")
	p := c.Param("page")
	user_id := c.GetInt64("userid")

	arch, err := s.repo.GetArchiveByID(ctx, archive_id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.ResponseError{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	page, _ := strconv.Atoi(p)
	page64 := int64(page)
	if page64 == 0 || page64 > arch.PageCount {
		c.JSON(http.StatusBadRequest, &models.ResponseError{
			Status:  "error",
			Message: "Invalid page value",
		})
		return
	}

	ok, err := s.repo.UserReadingProgressExists(ctx, repository.UserReadingProgressExistsParams{
		ArchiveID: arch.ID,
		UserID:    user_id,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.ResponseError{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	read := GetReadingProgress(page64, arch.PageCount)
	if ok.RowsAffected() == 0 {
		if err := s.repo.InsertReadingProgress(ctx, repository.InsertReadingProgressParams{
			ArchiveID: arch.ID,
			UserID:    user_id,
			Page:      read.Progress,
			State:     read.ReadingState,
			LastRead:  *read.LastRead,
		}); err != nil {
			c.JSON(http.StatusInternalServerError, &models.ResponseError{
				Status:  "error",
				Message: err.Error(),
			})
			return
		}
	} else {
		if _, err := s.repo.UpdateReadingProgress(ctx, repository.UpdateReadingProgressParams{
			ArchiveID: arch.ID,
			UserID:    user_id,
			Page:      &read.Progress,
			State:     &read.ReadingState,
			LastRead:  *read.LastRead,
		}); err != nil {
			c.JSON(http.StatusInternalServerError, &models.ResponseError{
				Status:  "error",
				Message: err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, &gin.H{
		"archive_id": archive_id,
		"userid":     user_id,
		"page":       read.Progress,
		"last_read":  read.LastRead,
		"read_state": read.ReadingState,
	})
}

func (s *Server) deleteReadingProgressHandler(c *gin.Context) {
	ctx := context.Background()

	archive_id := c.Param("id")
	user_id := c.GetInt64("userid")

	arch, err := s.repo.GetArchiveByID(ctx, archive_id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.ResponseError{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	if err := s.repo.DeleteReadingProgress(ctx, repository.DeleteReadingProgressParams{
		ArchiveID: arch.ID,
		UserID:    user_id,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, &models.ResponseError{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, &gin.H{
		"archive_id": archive_id,
		"userid":     user_id,
		"read_state": "unread",
	})
}
