package server

import (
	"Shoka/internal/archive"
	"Shoka/internal/config"
	"Shoka/internal/models"
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

func (s *Server) getAllCharacterHandler(c *gin.Context) {
	ctx := context.Background()

	characters, err := s.repo.GetAllCharacter(ctx)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, &models.ResponseError{
				Status:  "error",
				Message: config.ErrNoCharacter.Error(),
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

	c.JSON(http.StatusOK, characters)
}

func (s *Server) getArchiveByCharacterHandler(c *gin.Context) {
	// Get tag name from url parameter, makes it lowercase
	character := strings.ToLower(c.Param("character"))

	// Creates context
	ctx := context.Background()

	// Checks if tag exists
	exists, err := s.repo.CharacterExists(ctx, character)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	// If tag does not exists respond with 404 ErrTagNotFound
	if exists.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, &models.ResponseError{
			Status:  "error",
			Message: config.ErrCharacterNotFound.Error(),
		})
		return
	}

	// Get archives
	archives, err := archive.GetByCharacter(ctx, s.repo, character, s.log)
	if err != nil {
		// If no archives found respond with 404 ErrNoArchive
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, &models.ResponseError{
				Status:  "error",
				Message: config.ErrNoArchive.Error(),
			})
			return
			// Otherwise respond with 500 and error
		} else {
			c.JSON(http.StatusInternalServerError, &models.ResponseError{
				Status:  "error",
				Message: err.Error(),
			})
			return
		}
	}

	// Respond with 200 Success
	c.JSON(http.StatusOK, archives)
}
