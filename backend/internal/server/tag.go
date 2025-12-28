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

func (s *Server) getAllTagHandler(c *gin.Context) {
	ctx := context.Background()

	tags, err := s.repo.GetAllTags(ctx)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, &models.Response{
				Status:  "error",
				Message: config.ErrNoTags.Error(),
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

	c.JSON(http.StatusOK, tags)
}

func (s *Server) getArchiveByTagHandler(c *gin.Context) {
	// get tag param & pagination queries
	tag := strings.ToLower(c.Param("tag"))
	p := c.Query("page")
	ps := c.Query("size")

	ctx := context.Background()

	exists, err := s.repo.TagExists(ctx, tag)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	// If tag does not exists respond with 404 ErrTagNotFound
	if exists.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, &models.Response{
			Status:  "error",
			Message: config.ErrTagNotFound.Error(),
		})
		return
	}

	var uid uuid.UUID
	header := c.Request.Header.Get("Authorization")
	if header != "" {
		uid = util.GetUserFromRequest(c)
	}

	// Get archives

	if p == "" && ps == "" {
		archives, err := s.repo.GetArchiveByTag(ctx, repository.GetArchiveByTagParams{
			Name: tag,
			Uid:  uid,
		})
		if err != nil {
			// If no archives found respond with 404 ErrNoArchive
			if err == pgx.ErrNoRows {
				c.JSON(http.StatusNotFound, &models.Response{
					Status:  "error",
					Message: config.ErrNoArchive.Error(),
				})
				return
			} else {
				// Otherwise respond with 500 and error
				c.JSON(http.StatusInternalServerError, &models.Response{
					Status:  "error",
					Message: err.Error(),
				})
				return
			}
		}

		total, err := s.repo.TotalArchiveWithTag(ctx, tag)
		if err != nil {
			c.JSON(http.StatusInternalServerError, &models.Response{
				Status:  "error",
				Message: err.Error(),
			})
			return
		}
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

		archives, err := s.repo.GetArchiveByTagList(ctx, repository.GetArchiveByTagListParams{
			Name:   tag,
			Offset: (int32(page) - 1) * int32(pageSize),
			Limit:  int32(pageSize),
			Uid:    uid,
		})
		if err != nil {
			// If no archives found respond with 404 ErrNoArchive
			if err == pgx.ErrNoRows {
				c.JSON(http.StatusNotFound, &models.Response{
					Status:  "error",
					Message: config.ErrNoArchive.Error(),
				})
				return
			} else {
				// Otherwise respond with 500 and error
				c.JSON(http.StatusInternalServerError, &models.Response{
					Status:  "error",
					Message: err.Error(),
				})
				return
			}
		}
		total, err := s.repo.TotalArchiveWithTag(ctx, tag)
		if err != nil {
			c.JSON(http.StatusInternalServerError, &models.Response{
				Status:  "error",
				Message: err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"archives": archives,
			"total":    total,
		})
	}
}
