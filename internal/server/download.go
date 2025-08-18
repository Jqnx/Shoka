package server

import (
	"Shoka/internal/config"
	"Shoka/internal/downloader"
	"Shoka/internal/models"
	"Shoka/internal/repository"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type DownloadRequest struct {
	URL string `json:"url" binding:"required"`
}

func (s *Server) addDownloadHandler(c *gin.Context) {
	var req DownloadRequest
	if errPost := c.ShouldBind(&req); errPost != nil {
		var verr validator.ValidationErrors
		if errors.As(errPost, &verr) {
			c.JSON(http.StatusBadRequest, &models.ResponseFail{
				Status: "fail",
				Data:   config.Validate(verr),
			})
			return
		}

		c.JSON(http.StatusBadRequest, &models.ResponseFail{
			Status: "fail",
			Data:   errPost.Error(),
		})
		return
	}

	if _, err := url.ParseRequestURI(req.URL); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid URL"})
		return
	}

	download, err := s.dm.AddDownload(req.URL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.Response{
			Status:  "error",
			Message: err,
		})
		return
	}

	s.dm.BroadcastUpdate(download)
	c.JSON(http.StatusCreated, download)
}

func (s *Server) getAllDownloadsHandler(c *gin.Context) {
	ctx := context.Background()

	downloads, err := s.repo.GetAllDownloads(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.Response{
			Status:  "error",
			Message: err,
		})
		return
	}

	s.dm.EnrichWithActiveProgress(downloads)

	c.JSON(http.StatusOK, downloads)
}

func (s *Server) getDownloadHandler(c *gin.Context) {
	id := c.Param("id")
	downloadID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, &models.Response{
			Status:  "error",
			Message: "Invalid download ID",
		})
		return
	}

	ctx := context.Background()
	download, err := s.repo.GetDownload(ctx, downloadID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.Response{
			Status:  "error",
			Message: err,
		})
		return
	}

	// Enrich with current progress if active
	downloads := []repository.Download{download}
	s.dm.EnrichWithActiveProgress(downloads)

	c.JSON(http.StatusOK, downloads[0])
}

func (s *Server) getActiveDownloadHandler(c *gin.Context) {
	s.dm.ActiveMutex.RLock()
	defer s.dm.ActiveMutex.RUnlock()

	type Info struct {
		ID              string    `json:"id"`
		Progress        int32     `json:"progress"`
		Speed           int64     `json:"speed"`
		Downloaded      int64     `json:"downloaded"`
		TotalSize       int64     `json:"total_size"`
		StartTime       time.Time `json:"start_time"`
		CanResume       bool      `json:"can_resume"`
		ResumeSupported bool      `json:"resume_supported"`
	}

	activeInfo := make(map[string]Info)
	for id, activeDownload := range s.dm.ActiveDownloads {
		activeInfo[id] = Info{
			ID:              activeDownload.ID,
			Progress:        activeDownload.Progress,
			Speed:           activeDownload.DownloadSpeed,
			Downloaded:      activeDownload.BytesDownloaded,
			TotalSize:       activeDownload.TotalSize,
			StartTime:       activeDownload.StartTime,
			CanResume:       activeDownload.CanResume,
			ResumeSupported: activeDownload.ResumeSupported,
		}
	}

	c.JSON(http.StatusOK, activeInfo)
}

func (s *Server) pauseDownloadHandler(c *gin.Context) {
	id := c.Param("id")

	if err := s.dm.PauseDownload(id); err != nil {
		c.JSON(http.StatusInternalServerError, &models.Response{
			Status:  "error",
			Message: err,
		})
		return
	}

	c.JSON(http.StatusOK, &models.Response{
		Status:  "success",
		Message: "Download paused.",
	})
}

func (s *Server) resumeDownloadHandler(c *gin.Context) {
	ctx := context.Background()
	id := c.Param("id")

	downloadID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, &models.Response{
			Status:  "error",
			Message: "Invalid download id",
		})
		return
	}

	download, err := s.repo.GetDownload(ctx, downloadID)
	if err != nil {
		c.JSON(http.StatusNotFound, err)
		return
	}

	if download.Status != downloader.StatusPaused && download.Status != downloader.StatusFailed {
		c.JSON(http.StatusBadRequest, &models.Response{
			Status:  "error",
			Message: "Download can not be resumed from current state",
		})
		return
	}

	if err := s.dm.ResumeDownload(ctx, &download); err != nil {
		c.JSON(http.StatusInternalServerError, &models.Response{
			Status:  "error",
			Message: err,
		})
		return
	}

	c.JSON(http.StatusOK, &models.Response{
		Status:  "success",
		Message: "Download resumed",
	})
}

func (s *Server) deleteDownloadHandler(c *gin.Context) {
	id := c.Param("id")
	delFile, err := strconv.ParseBool(c.Query("file"))
	if err != nil {
		c.JSON(http.StatusBadRequest, &models.Response{
			Status:  "error",
			Message: "invalid bool value, must be true/false",
		})
		return
	}

	downloadID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, &models.Response{
			Status:  "error",
			Message: "Invalid download ID",
		})
		return
	}

	ctx := context.Background()
	download, err := s.repo.GetDownload(ctx, downloadID)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusInternalServerError, &models.Response{
				Status:  "error",
				Message: fmt.Sprintf("No download found with id: %v", downloadID),
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

	if err := s.dm.DeleteDownload(ctx, &download, delFile); err != nil {
		c.JSON(http.StatusInternalServerError, &models.Response{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": "Download deleted"})
}
