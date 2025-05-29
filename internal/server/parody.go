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

func (s *Server) getAllParodyHandler(c *gin.Context) {
	ctx := context.Background()

	parodies, err := s.repo.GetAllParodies(ctx)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, &models.ResponseError{
				Status:  "error",
				Message: config.ErrNoParody.Error(),
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

	c.JSON(http.StatusOK, parodies)
}

func (s *Server) getArchiveByParodyHandler(c *gin.Context) {
	// Get tag name from url parameter, makes it lowercase
	parody := strings.ToLower(c.Param("parody"))

	// Creates context
	ctx := context.Background()

	// Checks if tag exists
	exists, err := s.repo.ParodyExists(ctx, parody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	// If tag does not exists respond with 404 ErrTagNotFound
	if exists.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, &models.ResponseError{
			Status:  "error",
			Message: config.ErrParodyNotFound.Error(),
		})
		return
	}

	// Get archives
	archives, err := archive.GetByParody(ctx, s.repo, parody, s.log)
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
