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

func (s *Server) getAllCategoryHandler(c *gin.Context) {
	ctx := context.Background()

	categories, err := s.repo.GetAllCategory(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.Response{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, categories)
}

func (s *Server) getArchiveByCategoryHandler(c *gin.Context) {
	ctx := context.Background()

	category := strings.ToLower(c.Param("category"))

	page, _ := strconv.Atoi(c.Query("page"))
	if page == 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.Query("size"))
	if pageSize == 0 {
		pageSize = 10
	}

	order := util.GetSortOrderFromRequest(c)

	exists, err := s.repo.CategoryExists(ctx, &category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	if exists.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, &models.Response{
			Status:  "error",
			Message: config.ErrCharacterNotFound.Error(),
		})
		return
	}

	var uid uuid.UUID
	header := c.Request.Header.Get("Authorization")
	if header != "" {
		uid = util.GetUserFromRequest(c)
	}

	archives, err := s.repo.GetArchiveByCategoryList(ctx, repository.GetArchiveByCategoryListParams{
		Category: &category,
		Offset:   (int32(page) - 1) * int32(pageSize),
		Limit:    int32(pageSize),
		Uid:      uid,
		OrderBy:  order,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.Response{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}
	total, err := s.repo.TotalArchiveWithCategory(ctx, &category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.Response{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, &models.ArchiveListResponse[repository.GetArchiveByCategoryListRow]{
		Archives: archives,
		Count:    int(total),
	})
}
