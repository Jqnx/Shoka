package server

import (
	"Shoka/internal/models"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type ArchivePayload struct {
	Title     string   `json:"title" form:"title" binding:"required"`
	Summary   string   `json:"summary" form:"summary"`
	Tags      []string `json:"tags" form:"tags"`
	Artist    []string `json:"artist" form:"artist"`
	Parody    []string `json:"parody" form:"parody"`
	Character []string `json:"character" form:"character"`
	Language  string   `json:"language" form:"language" binding:"required,bcp47_language_tag"`
	Category  string   `json:"category" form:"category"`
	URL       []string `json:"url" form:"url"`
}

// Create
func (s *Server) createArchiveHandler(c *gin.Context) {
	var payload ArchivePayload

	if errPost := c.ShouldBind(&payload); errPost != nil {
		var verr validator.ValidationErrors
		if errors.As(errPost, &verr) {
			c.JSON(http.StatusBadRequest, gin.H{"errors": Validate(verr)})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{"errors": errPost.Error()})
		return

	}
	tags := s.tagsToStruct(payload.Tags)
	artists := s.artistToStruct(payload.Artist)
	parodies := s.parodyToStruct(payload.Parody)
	characters := s.characterToStruct(payload.Character)
	urls := s.urlToStruct(payload.URL)

	archive := &models.Archive{
		Title:     payload.Title,
		Summary:   payload.Summary,
		Tags:      tags,
		Artist:    artists,
		Parody:    parodies,
		Character: characters,
		Language:  strings.ToLower(payload.Language),
		Category:  strings.ToLower(payload.Category),
		URL:       urls,
	}

	if err := s.store.Archive.Create(c, archive); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"internal error": err.Error()})
		return
	}

	aid := archive.AID

	output, err := s.store.Archive.Get(aid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"internal error": err.Error()})
	}

	c.JSON(http.StatusCreated, gin.H{"created": output})
}

// Read

func (s *Server) getLastIDHandler(c *gin.Context) {
	lastid, err := s.store.Archive.GetLastID()
	fmt.Println(lastid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, lastid)
}

func (s *Server) getAllArchiveHandler(c *gin.Context) {
	archives, err := s.store.Archive.GetAll()
	// fmt.Println(archives)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, archives)
}

func (s *Server) getArchiveHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	archive, err := s.store.Archive.Get(id)
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
		var verr validator.ValidationErrors
		if errors.As(errPost, &verr) {
			c.JSON(http.StatusBadRequest, gin.H{"errors": Validate(verr)})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{"errors": errPost.Error()})
		return
	}

	tags := s.tagsToStruct(payload.Tags)
	artists := s.artistToStruct(payload.Artist)
	parodies := s.parodyToStruct(payload.Parody)
	characters := s.characterToStruct(payload.Character)
	urls := s.urlToStruct(payload.URL)
	idparam, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	id, err := s.store.Archive.GetID(idparam)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	archive := &models.Archive{
		ID:        id,
		Title:     payload.Title,
		Summary:   payload.Summary,
		Tags:      tags,
		Artist:    artists,
		Parody:    parodies,
		Character: characters,
		Language:  strings.ToLower(payload.Language),
		Category:  strings.ToLower(payload.Category),
		URL:       urls,
	}

	if err := s.store.Archive.Update(archive); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"internal error": err.Error()})
		return
	}

	output, err := s.store.Archive.Get(idparam)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"internal error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"updated": output})
}

// Delete

type archiveDeleted struct {
	ID uint
}

func (s *Server) deleteArchiveHandler(c *gin.Context) {
	idparam, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	id, err := s.store.Archive.GetID(idparam)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	archive := &models.Archive{
		ID: id,
	}

	if err := s.store.Archive.Delete(archive); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	output := &archiveDeleted{
		ID: archive.ID,
	}

	c.JSON(http.StatusOK, gin.H{"deleted": output})
}
