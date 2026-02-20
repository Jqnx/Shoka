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

func (s *Server) getAllParodyHandler(c *gin.Context) {
	ctx := context.Background()

	parodies, err := s.repo.GetAllParody(ctx)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, &models.Response{
				Status:  "error",
				Message: config.ErrNoParody.Error(),
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

	c.JSON(http.StatusOK, parodies)
}

func (s *Server) getArchiveByParodyHandler(c *gin.Context) {
	ctx := context.Background()
	parody := strings.ToLower(c.Param("parody"))
	page, _ := strconv.Atoi(c.Query("page"))
	if page == 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.Query("size"))
	if pageSize == 0 {
		pageSize = 10
	}
	order := util.GetSortOrderFromRequest(c)

	exists, err := s.repo.ParodyExists(ctx, parody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	if exists.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, &models.Response{
			Status:  "error",
			Message: config.ErrParodyNotFound.Error(),
		})
		return
	}

	var uid uuid.UUID
	header := c.Request.Header.Get("Authorization")
	if header != "" {
		uid = util.GetUserFromRequest(c)
	}

	archives, err := s.repo.GetArchiveByParodyList(ctx, repository.GetArchiveByParodyListParams{
		Name:    parody,
		Offset:  (int32(page) - 1) * int32(pageSize),
		Limit:   int32(pageSize),
		Uid:     uid,
		OrderBy: order,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.Response{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	total, err := s.repo.TotalArchiveWithParody(ctx, parody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.Response{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, &models.ArchiveListResponse[repository.GetArchiveByParodyListRow]{
		Archives: archives,
		Count:    int(total),
	})
}
