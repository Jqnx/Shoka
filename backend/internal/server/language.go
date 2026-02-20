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
)

func (s *Server) getAllLanguageHandler(c *gin.Context) {
	ctx := context.Background()

	languages, err := s.repo.GetAllLanguage(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.Response{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, languages)
}

func (s *Server) getArchiveByLanguageHandler(c *gin.Context) {
	ctx := context.Background()

	language := strings.ToLower(c.Param("language"))

	page, _ := strconv.Atoi(c.Query("page"))
	if page == 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.Query("size"))
	if pageSize == 0 {
		pageSize = 10
	}

	order := util.GetSortOrderFromRequest(c)

	exists, err := s.repo.LanguageExists(ctx, &language)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.Response{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	if exists.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, &models.Response{
			Status:  "error",
			Message: config.ErrLanguageNotFound.Error(),
		})
		return
	}

	var uid uuid.UUID
	header := c.Request.Header.Get("Authorization")
	if header != "" {
		uid = util.GetUserFromRequest(c)
	}

	archives, err := s.repo.GetArchiveByLanguageList(ctx, repository.GetArchiveByLanguageListParams{
		Uid:      uid,
		Language: &language,
		Offset:   (int32(page) - 1) * int32(pageSize),
		Limit:    int32(pageSize),
		OrderBy:  order,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.Response{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}
	total, err := s.repo.TotalArchiveWithLanguage(ctx, &language)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.Response{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, &models.ArchiveListResponse[repository.GetArchiveByLanguageListRow]{
		Archives: archives,
		Count:    int(total),
	})
}
