package server

import (
	"Shoka/internal/config"
	"Shoka/internal/models"
	"Shoka/internal/repository"
	"Shoka/internal/util"
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

func (s *Server) registerUser(c *gin.Context) {
	var payload models.UserPayload
	if errPost := c.ShouldBind(&payload); errPost != nil {
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
	ctx := context.Background()

	_, err := s.repo.GetUserByName(ctx, payload.Name)
	if err != nil {
		passHash, err := bcrypt.GenerateFromPassword([]byte(payload.Password), 10)
		if err != nil {
			c.JSON(http.StatusInternalServerError, &models.ResponseError{
				Status:  "error",
				Message: err.Error(),
			})
			return
		}

		_, err = s.repo.CreateUser(ctx, repository.CreateUserParams{
			Name:      payload.Name,
			Password:  string(passHash),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, &models.ResponseError{
				Status:  "error",
				Message: err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "User registered successfully!",
		})
	} else {
		c.JSON(http.StatusConflict, &models.ResponseError{
			Status:  "error",
			Message: "Username already exists.",
		})
		return
	}
}

func (s *Server) signInUser(c *gin.Context) {
	var payload models.UserPayload
	if errPost := c.ShouldBind(&payload); errPost != nil {
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

	ctx := context.Background()

	user, err := s.repo.GetUserByName(ctx, payload.Name)
	if err != nil {
		c.JSON(http.StatusNotFound, &models.ResponseError{
			Status:  "error",
			Message: "User not found.",
		})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(payload.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, &models.ResponseError{
			Status:  "error",
			Message: "Invalid password.",
		})
		return
	}

	token, err := util.GenerateToken(32)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.ResponseError{
			Status:  "error",
			Message: "Failed to generate token.",
		})
		return
	}

	expiry := time.Now().Add(time.Hour * 24 * 3)
	if err := s.repo.UpdateSession(ctx, repository.UpdateSessionParams{
		ID:            user.ID,
		Session:       &token,
		SessionExpiry: &expiry,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, &models.ResponseError{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":  token,
		"expiry": expiry,
	})
}

func (s *Server) getSession(c *gin.Context) {
	ctx := context.Background()
	username := c.GetString("username")

	user, err := s.repo.GetUserByName(ctx, username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.ResponseError{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, &models.UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		CreatedAt: user.CreatedAt,
	})
}

func (s *Server) signOutUser(c *gin.Context) {
	ctx := context.Background()
	username := c.GetString("username")

	user, err := s.repo.GetUserByName(ctx, username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.ResponseError{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	if err := s.repo.UpdateSession(ctx, repository.UpdateSessionParams{
		ID:            user.ID,
		Session:       nil,
		SessionExpiry: nil,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, &models.ResponseError{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "Successfully logged out.",
	})
}
