package server

import (
	"Shoka/internal/config"
	"Shoka/internal/group"
	"Shoka/internal/models"
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// TODO:

// Create
func (s *Server) createGroupHandler(c *gin.Context) {
	var payload models.GroupPayload
	// Binds payload
	if errPost := c.ShouldBind(&payload); errPost != nil {
		// Validates payload, returns 400 on error
		var verr validator.ValidationErrors
		if errors.As(errPost, &verr) {
			c.JSON(http.StatusBadRequest, &models.ResponseFail{
				Status: "fail",
				Data:   config.Validate(verr),
			})
			return
		}

		c.JSON(http.StatusBadRequest, &models.ResponseFail{
			Status: "fail",
			Data:   errPost.Error(),
		})
		return
	}

	// Cleans artist list of empty strings
	artists := config.ValidateGroupArtists(payload.Artists)
	payload.Artists = artists

	// Creates context
	ctx := context.Background()

	// Makes group name lowercase
	name := strings.ToLower(payload.Name)

	// Checks if group exists
	result, err := s.repo.GroupExists(ctx, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.ResponseError{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	// If group does not exist respond with 404 GroupNotFound error
	if result.RowsAffected() != 0 {
		c.JSON(http.StatusConflict, &models.ResponseError{
			Status:  "error",
			Message: config.ErrGroupNoDuplicates.Error(),
		})
		return
	}

	// Create group
	group, err := group.CreateTransaction(ctx, s.db, s.repo, &payload, s.log)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.ResponseError{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	// Respond with 200 Success
	c.JSON(http.StatusCreated, group)
}

// Get
func (s *Server) getGroupHandler(c *gin.Context) {
	// Gets name parameter from url, makes it lowercase
	name := strings.ToLower(c.Param("name"))

	// Creates context
	ctx := context.Background()

	// Get group
	group, err := group.Get(ctx, s.repo, name, s.log)
	if err != nil {
		// If group not found respond with 404 ErrGroupNotFound
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, &models.ResponseError{
				Status:  "error",
				Message: config.ErrGroupNotFound.Error(),
			})
			return
			// Otherwise respond with 500 and the error
		} else {
			c.JSON(http.StatusInternalServerError, &models.ResponseError{
				Status:  "error",
				Message: err.Error(),
			})
			return
		}
	}

	// Respond with 200 Success
	c.JSON(http.StatusOK, group)
}

func (s *Server) getAllGroupHandler(c *gin.Context) {
	// Create context
	ctx := context.Background()

	// Gets all groups
	groups, err := group.GetAll(ctx, s.repo, s.log)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.ResponseError{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	// If no groups found respond with 404 ErrNoGroup
	if len(*groups) == 0 {
		c.JSON(http.StatusNotFound, &models.ResponseError{
			Status:  "error",
			Message: config.ErrNoGroup.Error(),
		})
		return
	}

	// Respond with 200 Success
	c.JSON(http.StatusOK, groups)
}

// Update
func (s *Server) updateGroupHandler(c *gin.Context) {
	var payload models.GroupPayload
	// Binds payload
	if errPost := c.ShouldBind(&payload); errPost != nil {
		// Validates payload, returns 400 on error
		var verr validator.ValidationErrors
		if errors.As(errPost, &verr) {
			c.JSON(http.StatusBadRequest, &models.ResponseFail{
				Status: "fail",
				Data:   config.Validate(verr),
			})
			return
		}

		c.JSON(http.StatusBadRequest, &models.ResponseFail{
			Status: "fail",
			Data:   errPost.Error(),
		})
		return
	}

	// Cleans artist list of empty strings
	artists := config.ValidateGroupArtists(payload.Artists)
	payload.Artists = artists

	// Gets name parameter from url, makes it lowercase
	name := strings.ToLower(c.Param("name"))

	// Creates context
	ctx := context.Background()

	// Checks if group exists
	exists, err := s.repo.GroupExists(ctx, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	// If group does not exist respond with 404 GroupNotFound error
	if exists.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, &models.ResponseError{
			Status:  "error",
			Message: config.ErrGroupNotFound.Error(),
		})
		return
	}

	// Updates group
	result, err := group.UpdateTransaction(ctx,
		s.db,
		s.repo,
		&payload,
		name,
		s.log,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.ResponseError{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	// Responds with 200 success
	c.JSON(http.StatusOK, result)
}

// Delete
func (s *Server) deleteGroupHandler(c *gin.Context) {
	// Gets name parameter from url, makes it lowercase
	name := strings.ToLower(c.Param("name"))

	// Create context
	ctx := context.Background()

	// Check if group exists
	exists, err := s.repo.GroupExists(ctx, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.ResponseError{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	// If group does not exist respond with 404 GroupNotFound error
	if exists.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, &models.ResponseError{
			Status:  "error",
			Message: config.ErrGroupNotFound.Error(),
		})
		return
	}

	// Delete group
	if err := group.Delete(ctx, s.repo, name, s.log); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			// If err is foreign key constraint respond with 500 GroupHasMembers error
			if pgErr.Code == "23503" {
				c.JSON(http.StatusInternalServerError, &models.ResponseError{
					Status:  "error",
					Message: config.ErrGroupHasMembers.Error(),
				})
				return
				// Otherwise respond with 500 and the error
			} else {
				c.JSON(http.StatusInternalServerError, &models.ResponseError{
					Status:  "error",
					Message: err.Error(),
				})
				return
			}
		}
	}

	// Respond with 200 success
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}
