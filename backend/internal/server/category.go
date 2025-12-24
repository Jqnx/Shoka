package server

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"Shoka/internal/config"
	"Shoka/internal/models"
	"Shoka/internal/repository"
	"Shoka/internal/util"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *Server) getAllCategoryHandler(c *gin.Context) {
	ctx := context.Background()

	categories, err := s.repo.GetAllCategory(ctx)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, &models.Response{
				Status:  "error",
				Message: config.ErrNoCharacter.Error(),
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

	result := util.RemoveDuplicatesStrPointer(categories)

	c.JSON(http.StatusOK, result)
}

func (s *Server) getArchiveByCategoryHandler(c *gin.Context) {
	// Get tag name from url parameter, makes it lowercase
	category := strings.ToLower(c.Param("category"))
	p := c.Query("page")
	ps := c.Query("size")

	// Creates context
	ctx := context.Background()

	// Checks if tag exists
	exists, err := s.repo.CategoryExists(ctx, &category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	// If tag does not exists respond with 404 ErrTagNotFound
	if exists.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, &models.Response{
			Status:  "error",
			Message: config.ErrCharacterNotFound.Error(),
		})
		return
	}

	// TODO: Update to receive JWT
	var uid uuid.UUID
	header := c.Request.Header.Get("Authorization")
	if header != "" {
		uid = util.GetUserFromRequest(c)
	}

	// Get archives
	if p == "" && ps == "" {
		archives, err := s.repo.GetArchiveByCategory(ctx, repository.GetArchiveByCategoryParams{
			Category: &category,
			Uid:      uid,
		})
		if err != nil {
			// If no archives found respond with 404 ErrNoArchive
			if err == pgx.ErrNoRows {
				c.JSON(http.StatusNotFound, &models.Response{
					Status:  "error",
					Message: config.ErrNoArchive.Error(),
				})
				return
				// Otherwise respond with 500 and error
			} else {
				c.JSON(http.StatusInternalServerError, &models.Response{
					Status:  "error",
					Message: err.Error(),
				})
				return
			}
		}
		total, err := s.repo.TotalArchiveWithCategory(ctx, &category)
		if err != nil {
			c.JSON(http.StatusInternalServerError, &models.Response{
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
		archives, err := s.repo.GetArchiveByCategoryList(ctx, repository.GetArchiveByCategoryListParams{
			Category: &category,
			Offset:   (int32(page) - 1) * int32(pageSize),
			Limit:    int32(pageSize),
			Uid:      uid,
		})
		if err != nil {
			// If no archives found respond with 404 ErrNoArchive
			if err == pgx.ErrNoRows {
				c.JSON(http.StatusNotFound, &models.Response{
					Status:  "error",
					Message: config.ErrNoArchive.Error(),
				})
				return
				// Otherwise respond with 500 and error
			} else {
				c.JSON(http.StatusInternalServerError, &models.Response{
					Status:  "error",
					Message: err.Error(),
				})
				return
			}
		}
		total, err := s.repo.TotalArchiveWithCategory(ctx, &category)
		if err != nil {
			c.JSON(http.StatusInternalServerError, &models.Response{
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
