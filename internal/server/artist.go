package server

import (
	"Shoka/internal/artist"
	"Shoka/internal/models"
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// TODO:

// Create Handlers

// Creates an Artist
func (s *Server) createArtistHandler(c *gin.Context) {
	// Binding payload to struct
	var payload models.ArtistPayload
	if errPost := c.ShouldBind(&payload); errPost != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errPost.Error()})
		return
	}

	// Check if artist already exists
	result, err := s.repo.ArtistExists(c, strings.ToLower(payload.Name))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"internal error": err.Error()})
		return
	}
	if result.RowsAffected() != 0 {
		c.JSON(http.StatusConflict, gin.H{"internal error": ErrArtistNoDuplicates.Error()})
		return
	}

	// Create Artist
	payload.Name = strings.ToLower(payload.Name)
	ctx := context.Background()
	ctx = s.log.Logger.WithContext(ctx)
	artist, err := artist.CreateTransaction(ctx, s.db, s.repo, &payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"internal error": err.Error()})
		return
	}

	// Return JSON to Client
	c.JSON(http.StatusCreated, gin.H{"created": artist})
}

// Read Handlers

// Gets All Artists
func (s *Server) getAllArtistHandler(c *gin.Context) {
	ctx := context.Background()
	ctx = s.log.Logger.WithContext(ctx)
	artists, err := artist.GetAll(ctx, s.repo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, artists)
}

// Gets an individual Artist
func (s *Server) getArtistHandler(c *gin.Context) {
	ctx := context.Background()
	ctx = s.log.Logger.WithContext(ctx)
	artist, err := artist.Get(ctx, s.repo, c.Param("name"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, artist)
}
