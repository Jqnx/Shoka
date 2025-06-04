package server

import (
	"Shoka/internal/config"
	"Shoka/internal/models"
	"Shoka/internal/repository"
	"Shoka/internal/util"
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

func (s *Server) getAllLanguageHandler(c *gin.Context) {
	ctx := context.Background()

	languages, err := s.repo.GetAllLanguage(ctx)
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

	result := util.RemoveDuplicatesStrPointer(languages)

	c.JSON(http.StatusOK, result)
}

func (s *Server) getArchiveByLanguageHandler(c *gin.Context) {
	// Get tag name from url parameter, makes it lowercase
	language := strings.ToLower(c.Param("language"))
	p := c.Query("page")
	ps := c.Query("size")

	// Creates context
	ctx := context.Background()

	// Checks if tag exists
	exists, err := s.repo.LanguageExists(ctx, &language)
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
	if p == "" && ps == "" {
		archives, err := s.repo.GetArchivesByLanguage(ctx, &language)
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
		total, err := s.repo.TotalArchivesWithLanguage(ctx, &language)
		if err != nil {
			c.JSON(http.StatusInternalServerError, &models.ResponseError{
				Status:  "error",
				Message: err.Error(),
			})
			return
		}

		// Respond with 200 Success
		c.JSON(http.StatusOK, gin.H{
			"archives": archives,
			"total":    total,
		})
	} else {

		page, _ := strconv.Atoi(p)
		if page == 0 {
			page = 1
		}
		pageSize, _ := strconv.Atoi(ps)
		if pageSize == 0 {
			pageSize = 10
		}
		archives, err := s.repo.GetArchivesByLanguageList(ctx, repository.GetArchivesByLanguageListParams{
			Language: &language,
			Offset:   (int32(page) - 1) * int32(pageSize),
			Limit:    int32(pageSize),
		})
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
		total, err := s.repo.TotalArchivesWithLanguage(ctx, &language)
		if err != nil {
			c.JSON(http.StatusInternalServerError, &models.ResponseError{
				Status:  "error",
				Message: err.Error(),
			})
			return
		}

		// Respond with 200 Success
		c.JSON(http.StatusOK, gin.H{
			"archives": archives,
			"total":    total,
		})
	}
}
