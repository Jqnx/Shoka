package server

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"Shoka/internal/models"
	"Shoka/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ReadingProgress struct {
	Status   string
	Progress int16
	LastRead *time.Time
}

func GetReadingProgress(page, max int16) *ReadingProgress {
	now := time.Now()
	switch {
	case page == 0:
		return &ReadingProgress{
			Status:   "unread",
			Progress: 0,
			LastRead: nil,
		}
	case page > 0 && page < max:
		return &ReadingProgress{
			Status:   "reading",
			Progress: page,
			LastRead: &now,
		}
	case page == max:
		return &ReadingProgress{
			Status:   "finished",
			Progress: page,
			LastRead: &now,
		}
	}
	return nil
}

func (s *Server) getRecentlyReadHandler(c *gin.Context) {
	ctx := context.Background()

	user_id, _ := uuid.Parse(c.GetString("userid"))

	arch, err := s.repo.GetRecentlyReadArchives(ctx, user_id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.Response{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, &gin.H{
		"archives": arch,
		"total":    len(arch),
	})
}

func (s *Server) updateReadingProgressHandler(c *gin.Context) {
	ctx := context.Background()

	archive_id := c.Param("id")
	p := c.Param("page")
	user_id, _ := uuid.Parse(c.GetString("userid"))

	arch, err := s.repo.GetArchiveByID(ctx, archive_id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.Response{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	page, _ := strconv.Atoi(p)
	page16 := int16(page)
	if page16 == 0 || page16 > arch.PageCount {
		c.JSON(http.StatusBadRequest, &models.Response{
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
		c.JSON(http.StatusInternalServerError, &models.Response{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	read := GetReadingProgress(page16, arch.PageCount)
	if ok.RowsAffected() == 0 {
		if err := s.repo.InsertReadingProgress(ctx, repository.InsertReadingProgressParams{
			ArchiveID: arch.ID,
			UserID:    user_id,
			Page:      read.Progress,
			Status:    read.Status,
			LastRead:  *read.LastRead,
		}); err != nil {
			c.JSON(http.StatusInternalServerError, &models.Response{
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
			Status:    &read.Status,
			LastRead:  *read.LastRead,
		}); err != nil {
			c.JSON(http.StatusInternalServerError, &models.Response{
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
		"read_state": read.Status,
	})
}

func (s *Server) deleteReadingProgressHandler(c *gin.Context) {
	ctx := context.Background()

	archive_id := c.Param("id")
	user_id, _ := uuid.Parse(c.GetString("userid"))

	arch, err := s.repo.GetArchiveByID(ctx, archive_id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.Response{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	if err := s.repo.DeleteReadingProgress(ctx, repository.DeleteReadingProgressParams{
		ArchiveID: arch.ID,
		UserID:    user_id,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, &models.Response{
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
