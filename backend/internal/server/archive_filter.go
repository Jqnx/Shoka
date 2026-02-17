package server

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"Shoka/internal/filter"
	"Shoka/internal/models"
	"Shoka/internal/repository"
	"Shoka/internal/util"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TODO: Add metadata to individual archives, so basically the normal get for 1 archive but for a list
func (s *Server) getArchiveFilterHandler(c *gin.Context) {
	ctx := context.Background()
	filters, hasFilter := filter.GetFromQuery(c)

	var uid uuid.UUID
	header := c.Request.Header.Get("Authorization")
	if header != "" {
		uid = util.GetUserFromRequest(c)
	}

	order := fmt.Sprintf("%s_%s", c.Query("sortby"), c.Query("sortdir"))
	page, _ := strconv.Atoi(c.Query("page"))
	if page == 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.Query("size"))
	if pageSize == 0 {
		pageSize = 10
	}

	// Query with filters if query is not empty
	if hasFilter {
		archives, err := filter.MatchAndGet(s.repo, filter.MatchAndGetParams{
			Filters:  *filters,
			Page:     page,
			PageSize: pageSize,
			Order:    order,
			UserID:   uid,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, &models.Response{
				Status:  "error",
				Message: err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, archives)
		return
	} else {
		count, err := s.repo.CountArchives(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, &models.Response{
				Status:  "error",
				Message: err.Error(),
			})
			return
		}
		archives, err := s.repo.GetArchiveSortList(ctx, repository.GetArchiveSortListParams{
			Uid:     uid,
			OrderBy: order,
			Limit:   int32(pageSize),
			Offset:  (int32(page) - 1) * int32(pageSize),
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, &models.Response{
				Status:  "error",
				Message: err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, &models.ArchiveListResponse[repository.GetArchiveSortListRow]{
			Archives: archives,
			Count:    int(count),
		})
		return
	}
}

func (s *Server) getAllFiltersHandler(c *gin.Context) {
	ctx := context.Background()

	tags, err := s.repo.GetAllTags(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.Response{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}
	artists, err := s.repo.GetAllArtists(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.Response{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}
	characters, err := s.repo.GetAllCharacter(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.Response{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}
	parodies, err := s.repo.GetAllParody(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.Response{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}
	languages, err := s.repo.GetAllLanguage(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.Response{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}
	categories, err := s.repo.GetAllCategory(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.Response{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, &gin.H{
		"tags":       tags,
		"artists":    artists,
		"characters": characters,
		"parodies":   parodies,
		"languages":  languages,
		"categories": categories,
	})
}
