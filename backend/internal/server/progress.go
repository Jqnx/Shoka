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

func GetReadingProgress(page, max int16) *models.ReadingProgress {
	now := time.Now()
	switch {
	case page == 0:
		return &models.ReadingProgress{
			Status:   models.StateUnread,
			Progress: 0,
			LastRead: nil,
		}
	case page > 0 && page < max:
		return &models.ReadingProgress{
			Status:   models.StateReading,
			Progress: page,
			LastRead: &now,
		}
	case page == max:
		return &models.ReadingProgress{
			Status:   models.StateFinished,
			Progress: page,
			LastRead: &now,
		}
	}
	return nil
}

func (s *Server) getRecentlyReadHandler(c *gin.Context) {
	ctx := context.Background()

	userID, _ := uuid.Parse(c.GetString("userid"))

	arch, err := s.repo.GetRecentlyReadArchives(ctx, userID)
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

	archiveID := c.Param("id")
	p := c.Param("page")
	userID, _ := uuid.Parse(c.GetString("userid"))

	arch, err := s.repo.GetArchiveByID(ctx, archiveID)
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
		UserID:    userID,
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
			UserID:    userID,
			Page:      read.Progress,
			Status:    read.Status.String(),
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
			UserID:    userID,
			Page:      read.Progress,
			Status:    read.Status.String(),
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
		"archive_id": archiveID,
		"userid":     userID,
		"page":       read.Progress,
		"last_read":  read.LastRead,
		"read_state": read.Status.String(),
	})
}

func (s *Server) deleteReadingProgressHandler(c *gin.Context) {
	ctx := context.Background()

	archiveID := c.Param("id")
	userID, _ := uuid.Parse(c.GetString("userid"))

	arch, err := s.repo.GetArchiveByID(ctx, archiveID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.Response{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	if err := s.repo.DeleteReadingProgress(ctx, repository.DeleteReadingProgressParams{
		ArchiveID: arch.ID,
		UserID:    userID,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, &models.Response{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, &gin.H{
		"archive_id": archiveID,
		"userid":     userID,
		"read_state": models.StateUnread.String(),
	})
}
