package server

import (
	"Shoka/internal/artist"
	"Shoka/internal/config"
	"Shoka/internal/models"
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"
)

// TODO:

// Create Handlers

// Creates an Artist
func (s *Server) createArtistHandler(c *gin.Context) {
	// Binding payload to struct
	var payload models.ArtistPayload
	if errPost := c.ShouldBind(&payload); errPost != nil {
		// Validate payload
		var verr validator.ValidationErrors
		if errors.As(errPost, &verr) {
			c.JSON(http.StatusBadRequest, models.ResponseFail{
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

	// Check if artist already exists
	result, err := s.repo.ArtistExists(c, strings.ToLower(payload.Name))
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.Response{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}
	if result.RowsAffected() != 0 {
		c.JSON(http.StatusConflict, &models.Response{
			Status:  "error",
			Message: config.ErrArchiveNoDuplicates.Error(),
		})
		return
	}

	// Create Artist
	payload.Name = strings.ToLower(payload.Name)
	ctx := context.Background()
	res, err := artist.CreateTransaction(ctx, s.db, s.repo, &payload, s.log)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.Response{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	// Return JSON to Client
	c.JSON(http.StatusCreated, res)
}

// Read Handlers

// Gets All Artists
func (s *Server) getAllArtistHandler(c *gin.Context) {
	ctx := context.Background()
	artists, err := artist.GetAll(ctx, s.repo, s.log)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.Response{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	if len(*artists) == 0 {
		c.JSON(http.StatusNotFound, &models.Response{
			Status:  "error",
			Message: config.ErrNoArtist.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, artists)
}

// Gets an individual Artist
func (s *Server) getArtistHandler(c *gin.Context) {
	ctx := context.Background()
	artist, err := artist.Get(ctx, s.repo, c.Param("name"), s.log)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, &models.Response{
				Status:  "error",
				Message: config.ErrArtistNotFound.Error(),
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

	c.JSON(http.StatusOK, artist)
}

// Update
func (s *Server) updateArtistHandler(c *gin.Context) {
	var payload models.ArtistPayload
	if errPost := c.ShouldBind(&payload); errPost != nil {
		// Validate payload
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

	name := strings.ToLower(c.Param("name"))

	ctx := context.Background()
	exists, err := s.repo.ArtistExists(ctx, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	if exists.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, &models.Response{
			Status:  "error",
			Message: config.ErrArtistNotFound.Error(),
		})
		return
	}

	result, err := artist.UpdateTransaction(ctx,
		s.db,
		s.repo,
		&payload,
		name,
		s.log,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	c.JSON(http.StatusOK, result)
}

// Delete
func (s *Server) deleteArtistHandler(c *gin.Context) {
	name := strings.ToLower(c.Param("name"))

	ctx := context.Background()
	if err := artist.DeleteTransaction(ctx, s.db, s.repo, name, s.log); err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, &models.Response{
				Status:  "error",
				Message: config.ErrArtistNotFound.Error(),
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

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}
