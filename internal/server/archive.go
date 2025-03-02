package server

import (
	"Shoka/internal/models"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ArchivePayload struct {
	Title     string             `json:"title" form:"title" binding:"required"`
	Summary   string             `json:"summary" form:"summary"`
	Tags      []models.Tag       `json:"tags" form:"tags"`
	Artist    []models.Artist    `json:"artist" form:"artist"`
	Parody    []models.Parody    `json:"parody" form:"parody"`
	Character []models.Character `json:"character" form:"character"`
	Language  string             `json:"language" form:"language"`
	Category  string             `json:"category" form:"category"`
	URL       []models.URL       `json:"url" form:"url"`
}

// Create

func (s *Server) createArchiveHandler(c *gin.Context) {
	var payload ArchivePayload

	if errPost := c.ShouldBind(&payload); errPost != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errPost.Error()})
		return
	}

	archive := &models.Archive{
		Title:     payload.Title,
		Summary:   payload.Summary,
		Tags:      payload.Tags,
		Artist:    payload.Artist,
		Parody:    payload.Parody,
		Character: payload.Character,
		Language:  payload.Language,
		Category:  payload.Category,
		URL:       payload.URL,
	}

	if err := s.store.Archive.Create(c, archive); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"internal error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"created": archive})
}

// Read

func (s *Server) getLastIDHandler(c *gin.Context) {
	err, lastid := s.store.Archive.GetLastID()
	fmt.Println(lastid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, lastid)
}

func (s *Server) getAllArchiveHandler(c *gin.Context) {
	err, archives := s.store.Archive.GetAll()
	// fmt.Println(archives)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, archives)
}

func (s *Server) getArchiveHandler(c *gin.Context) {
	id := c.Param("id")
	err, archive := s.store.Archive.Get(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, archive)
}

// Update

func (s *Server) updateArchiveHandler(c *gin.Context) {
	var payload ArchivePayload

	if errPost := c.ShouldBind(&payload); errPost != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errPost.Error()})
		return
	}

	archive := &models.Archive{
		Title:     payload.Title,
		Summary:   payload.Summary,
		Tags:      payload.Tags,
		Artist:    payload.Artist,
		Parody:    payload.Parody,
		Character: payload.Character,
		Language:  payload.Language,
		Category:  payload.Category,
		URL:       payload.URL,
	}

	if err := s.store.Archive.Create(c, archive); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"internal error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"created": archive})
}
