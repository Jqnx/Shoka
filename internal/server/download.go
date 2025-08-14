package server

import (
	"Shoka/internal/config"
	"Shoka/internal/models"
	"Shoka/internal/repository"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

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
		c.JSON(http.StatusInternalServerError, &models.ResponseError{
			Status:  "error",
			Message: err,
		})
		return
	}

	c.JSON(http.StatusCreated, download)
}

func (s *Server) getAllDownloadsHandler(c *gin.Context) {
	ctx := context.Background()

	downloads, err := s.repo.GetAllDownloads(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.ResponseError{
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
		c.JSON(http.StatusBadRequest, &models.ResponseError{
			Status:  "error",
			Message: "Invalid download ID",
		})
		return
	}

	ctx := context.Background()
	download, err := s.repo.GetDownload(ctx, downloadID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.ResponseError{
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

func (s *Server) deleteDownloadHandler(c *gin.Context) {
	id := c.Param("id")

	downloadID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, &models.ResponseError{
			Status:  "error",
			Message: "Invalid download ID",
		})
		return
	}

	ctx := context.Background()
	download, err := s.repo.GetDownload(ctx, downloadID)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusInternalServerError, &models.ResponseError{
				Status:  "error",
				Message: fmt.Sprintf("No download found with id: %v", downloadID),
			})
			return
		} else {
			c.JSON(http.StatusInternalServerError, &models.ResponseError{
				Status:  "error",
				Message: err.Error(),
			})
			return
		}
	}

	if err := s.dm.DeleteDownload(ctx, &download); err != nil {
		c.JSON(http.StatusInternalServerError, &models.ResponseError{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": "Download deleted"})
}
