package server

import (
	"Shoka/internal/archive"
	"Shoka/internal/models"
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// TODO:
// Queries for creating metadata
// Queries for getting archive with metadata
// Pretty print json response

// Create
func (s *Server) createArchiveHandler(c *gin.Context) {
	var payload models.ArchivePayload

	if errPost := c.ShouldBind(&payload); errPost != nil {
		var verr validator.ValidationErrors
		if errors.As(errPost, &verr) {
			c.JSON(http.StatusBadRequest, gin.H{"errors": Validate(verr)})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{"errors": errPost.Error()})
		return

	}
	// tags := s.tagsToStruct(payload.Tags)
	// artists := s.artistToStruct(payload.Artist)
	// parodies := s.parodyToStruct(payload.Parody)
	// characters := s.characterToStruct(payload.Character)
	// urls := s.urlToStruct(payload.URL)

	//archive := &models.Archive{
	//	Title:     payload.Title,
	//	Summary:   payload.Summary,
	//	Tags:      tags,
	//	Artist:    artists,
	//	Parody:    parodies,
	//	Character: characters,
	//	Language:  strings.ToLower(payload.Language),
	//	Category:  strings.ToLower(payload.Category),
	//	URL:       urls,
	//	FilePath:  payload.FilePath,
	//}

	//if s.store.Archive.TitleExists(archive) {
	//	c.JSON(http.StatusConflict, gin.H{"internal error": ErrArchiveNoDuplicates.Error()})
	//	return
	//}

	//if err := s.store.Archive.Create(archive); err != nil {
	//	c.JSON(http.StatusInternalServerError, gin.H{"internal error": err.Error()})
	//	return
	//}

	ctx := context.Background()
	output, err := archive.CreateTransaction(ctx, s.db, s.repo, &payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"internal error": err.Error()})
		return
	}

	// aid := archive.AID

	// output, err := s.store.Archive.Get(aid)
	//if err != nil {
	//	c.JSON(http.StatusInternalServerError, gin.H{"internal error": err.Error()})
	//}

	c.JSON(http.StatusCreated, gin.H{"created": output})
}

// Read

func (s *Server) getLastIDHandler(c *gin.Context) {
	lastid, err := s.repo.GetArchiveLastAID(c)
	fmt.Println(lastid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, lastid)
}

func (s *Server) getAllArchiveHandler(c *gin.Context) {
	// archives, err := s.store.Archive.GetAll()
	archives, err := s.repo.GetAllArchives(c)
	fmt.Println(archives)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, archives)
}

func (s *Server) getArchiveHandler(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	archive, err := s.repo.GetArchiveByAID(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, archive)
}

// Update
func (s *Server) updateArchiveHandler(c *gin.Context) {
	var payload models.ArchivePayload

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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	id := s.store.Archive.GetID(idparam)
	if id == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": ErrArchiveNotFound.Error()})
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
	ID int
}

func (s *Server) deleteArchiveHandler(c *gin.Context) {
	idparam, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	id := s.store.Archive.GetID(idparam)
	if id == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": ErrArchiveNotFound.Error()})
		return
	}

	archive := &models.Archive{
		ID: id,
	}

	err = s.store.Archive.Delete(archive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	output := &archiveDeleted{
		ID: archive.ID,
	}

	c.JSON(http.StatusOK, gin.H{"deleted": output})
}
