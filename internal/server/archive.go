package server

import (
	"Shoka/internal/archive"
	"Shoka/internal/config"
	"Shoka/internal/models"
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"
)

// TODO:
// Print proper error responses instead of just printing errors directly

// Create
func (s *Server) createArchiveHandler(c *gin.Context) {
	var payload models.ArchivePayload
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
	result, err := s.repo.ArchiveExists(ctx, payload.Title)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.ResponseError{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}
	if result.RowsAffected() != 0 {
		c.JSON(http.StatusConflict, &models.ResponseError{
			Status:  "error",
			Message: config.ErrArchiveNoDuplicates.Error(),
		})
		return
	}

	archive, err := archive.CreateTransaction(ctx, s.db, s.repo, &payload, s.log)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.ResponseError{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, &models.ResponseSuccess{
		Status: "success",
		Data:   archive,
	})
}

// Read

//func (s *Server) getLastIDHandler(c *gin.Context) {
//	lastid, err := s.repo.GetArchiveLastAID(c)
//	if err != nil {
//		s.log.Error().AnErr("getLastIDHandler", err).Send()
//		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
//		return
//	}
//
//	c.JSON(http.StatusOK, lastid)
//}

func (s *Server) getAllArchiveHandler(c *gin.Context) {
	ctx := context.Background()

	archives, err := archive.GetAll(ctx, s.repo, s.log)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.ResponseError{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	if len(*archives) == 0 {
		c.JSON(http.StatusNotFound, &models.ResponseError{
			Status:  "error",
			Message: config.ErrNoArchive.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, &models.ResponseSuccess{
		Status: "success",
		Data:   archives,
	})
}

func (s *Server) getAllTagHandler(c *gin.Context) {
	ctx := context.Background()

	tags, err := s.repo.GetAllTags(ctx)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, &models.ResponseError{
				Status:  "error",
				Message: config.ErrNoTags.Error(),
			})
			return
		} else {
			c.JSON(http.StatusInternalServerError, &models.ResponseError{
				Status:  "error",
				Message: err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, &models.ResponseSuccess{
		Status: "success",
		Data:   tags,
	})
}

func (s *Server) getAllCharacterHandler(c *gin.Context) {
	ctx := context.Background()

	characters, err := s.repo.GetAllCharacter(ctx)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, &models.ResponseError{
				Status:  "error",
				Message: config.ErrNoCharacter.Error(),
			})
			return
		} else {
			c.JSON(http.StatusInternalServerError, &models.ResponseError{
				Status:  "error",
				Message: err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, &models.ResponseSuccess{
		Status: "success",
		Data:   characters,
	})
}

func (s *Server) getAllParodyHandler(c *gin.Context) {
	ctx := context.Background()

	parodies, err := s.repo.GetAllParodies(ctx)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, &models.ResponseError{
				Status:  "error",
				Message: config.ErrNoParody.Error(),
			})
			return
		} else {
			c.JSON(http.StatusInternalServerError, &models.ResponseError{
				Status:  "error",
				Message: err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, &models.ResponseSuccess{
		Status: "success",
		Data:   parodies,
	})
}

func (s *Server) getArchiveHandler(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.ResponseError{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	ctx := context.Background()
	archive, err := archive.Get(ctx, s.repo, id, s.log)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, &models.ResponseError{
				Status:  "error",
				Message: config.ErrArchiveNotFound.Error(),
			})
			return
		} else {
			c.JSON(http.StatusInternalServerError, &models.ResponseError{
				Status:  "error",
				Message: err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, &models.ResponseSuccess{
		Status: "success",
		Data:   archive,
	})
}

func (s *Server) getArchiveByTagHandler(c *gin.Context) {
	// Get tag name from url parameter, makes it lowercase
	tag := strings.ToLower(c.Param("tag"))

	// Creates context
	ctx := context.Background()

	// Checks if tag exists
	exists, err := s.repo.TagExists(ctx, tag)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	// If tag does not exists respond with 404 ErrTagNotFound
	if exists.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, &models.ResponseError{
			Status:  "error",
			Message: config.ErrTagNotFound.Error(),
		})
		return
	}

	// Get archives
	archives, err := archive.GetByTag(ctx, s.repo, tag, s.log)
	if err != nil {
		// If no archives found respond with 404 ErrNoArchive
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, &models.ResponseError{
				Status:  "error",
				Message: config.ErrNoArchive.Error(),
			})
			return
			// Otherwise respond with 500 and error
		} else {
			c.JSON(http.StatusInternalServerError, &models.ResponseError{
				Status:  "error",
				Message: err.Error(),
			})
			return
		}
	}

	// Respond with 200 Success
	c.JSON(http.StatusOK, &models.ResponseSuccess{
		Status: "success",
		Data:   archives,
	})
}

func (s *Server) getArchiveByCharacterHandler(c *gin.Context) {
	// Get tag name from url parameter, makes it lowercase
	character := strings.ToLower(c.Param("character"))

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
		c.JSON(http.StatusNotFound, &models.ResponseError{
			Status:  "error",
			Message: config.ErrCharacterNotFound.Error(),
		})
		return
	}

	// Get archives
	archives, err := archive.GetByCharacter(ctx, s.repo, character, s.log)
	if err != nil {
		// If no archives found respond with 404 ErrNoArchive
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, &models.ResponseError{
				Status:  "error",
				Message: config.ErrNoArchive.Error(),
			})
			return
			// Otherwise respond with 500 and error
		} else {
			c.JSON(http.StatusInternalServerError, &models.ResponseError{
				Status:  "error",
				Message: err.Error(),
			})
			return
		}
	}

	// Respond with 200 Success
	c.JSON(http.StatusOK, &models.ResponseSuccess{
		Status: "success",
		Data:   archives,
	})
}

func (s *Server) getArchiveByParodyHandler(c *gin.Context) {
	// Get tag name from url parameter, makes it lowercase
	parody := strings.ToLower(c.Param("parody"))

	// Creates context
	ctx := context.Background()

	// Checks if tag exists
	exists, err := s.repo.ParodyExists(ctx, parody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	// If tag does not exists respond with 404 ErrTagNotFound
	if exists.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, &models.ResponseError{
			Status:  "error",
			Message: config.ErrParodyNotFound.Error(),
		})
		return
	}

	// Get archives
	archives, err := archive.GetByParody(ctx, s.repo, parody, s.log)
	if err != nil {
		// If no archives found respond with 404 ErrNoArchive
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, &models.ResponseError{
				Status:  "error",
				Message: config.ErrNoArchive.Error(),
			})
			return
			// Otherwise respond with 500 and error
		} else {
			c.JSON(http.StatusInternalServerError, &models.ResponseError{
				Status:  "error",
				Message: err.Error(),
			})
			return
		}
	}

	// Respond with 200 Success
	c.JSON(http.StatusOK, &models.ResponseSuccess{
		Status: "success",
		Data:   archives,
	})
}

// Update
func (s *Server) updateArchiveHandler(c *gin.Context) {
	var payload models.ArchivePayload
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

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.ResponseError{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	ctx := context.Background()
	exists, err := s.repo.ArchiveAIDExists(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	if exists.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, &models.ResponseError{
			Status:  "error",
			Message: config.ErrArchiveNotFound.Error(),
		})
		return
	}

	result, err := archive.UpdateTransaction(ctx,
		s.db,
		s.repo,
		&payload,
		id,
		s.log,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.ResponseError{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, &models.ResponseSuccess{
		Status: "success",
		Data:   result,
	})
}

// Delete

func (s *Server) deleteArchiveHandler(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.ResponseError{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	ctx := context.Background()
	if err := archive.DeleteTransaction(ctx, s.db, s.repo, id, s.log); err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, &models.ResponseError{
				Status:  "error",
				Message: config.ErrArchiveNotFound.Error(),
			})
			return
		} else {
			c.JSON(http.StatusInternalServerError, &models.ResponseError{
				Status:  "error",
				Message: err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, &models.ResponseSuccess{
		Status: "success",
	})
}
