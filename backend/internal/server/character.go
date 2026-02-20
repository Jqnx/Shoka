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

func (s *Server) getAllCharacterHandler(c *gin.Context) {
	ctx := context.Background()

	characters, err := s.repo.GetAllCharacter(ctx)
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

	c.JSON(http.StatusOK, characters)
}

func (s *Server) getArchiveByCharacterHandler(c *gin.Context) {
	ctx := context.Background()
	character := strings.ToLower(c.Param("character"))
	page, _ := strconv.Atoi(c.Query("page"))
	if page == 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.Query("size"))
	if pageSize == 0 {
		pageSize = 10
	}

	order := util.GetSortOrderFromRequest(c)

	exists, err := s.repo.CharacterExists(ctx, character)
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

	archives, err := s.repo.GetArchiveByCharacterList(ctx, repository.GetArchiveByCharacterListParams{
		Name:    character,
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
	total, err := s.repo.TotalArchiveWithCharacter(ctx, character)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.Response{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, &models.ArchiveListResponse[repository.GetArchiveByCharacterListRow]{
		Archives: archives,
		Count:    int(total),
	})
}
