package server

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"Shoka/internal/config"
	"Shoka/internal/models"
	"Shoka/internal/repository"

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
	// Get tag name from url parameter, makes it lowercase
	character := strings.ToLower(c.Param("character"))
	p := c.Query("page")
	ps := c.Query("size")

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
		c.JSON(http.StatusNotFound, &models.Response{
			Status:  "error",
			Message: config.ErrCharacterNotFound.Error(),
		})
		return
	}

	// TODO: Update to receive JWT
	var uid uuid.UUID
	/*
		header := c.Request.Header.Get("Authorization")
		if header != "" {
			token := util.GetAuthTokenFromHeader(header)
			user, err := s.repo.GetUserByToken(ctx, token)
			if err != nil {
				c.JSON(http.StatusUnauthorized, &models.Response{
					Status:  "error",
					Message: "Unauthorized",
				})
				return
			}
			uid = user.ID
		}
	*/

	// Get archives
	if p == "" && ps == "" {
		archives, err := s.repo.GetArchiveByCharacter(ctx, repository.GetArchiveByCharacterParams{
			Name: character,
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
				// Otherwise respond with 500 and error
			} else {
				c.JSON(http.StatusInternalServerError, &models.Response{
					Status:  "error",
					Message: err.Error(),
				})
				return
			}
		}
		total, err := s.repo.TotalArchiveWithCharacter(ctx, character)
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
		archives, err := s.repo.GetArchiveByCharacterList(ctx, repository.GetArchiveByCharacterListParams{
			Name:   character,
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
				// Otherwise respond with 500 and error
			} else {
				c.JSON(http.StatusInternalServerError, &models.Response{
					Status:  "error",
					Message: err.Error(),
				})
				return
			}
		}
		total, err := s.repo.TotalArchiveWithCharacter(ctx, character)
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
