package server

import (
	"Shoka/internal/archive"
	"Shoka/internal/config"
	"Shoka/internal/fsutil"
	"Shoka/internal/models"
	"Shoka/internal/repository"
	"context"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"
)

// TODO: Print proper error responses instead of just printing errors directly
// TODO: Update handlers to use Metadata builder instead

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

	// TODO: Update handlers to use Metadata builder instead
	archive, err := archive.CreateTransaction(ctx, s.db, s.repo, &payload, s.log)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.ResponseError{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, archive)
}

// Read

func (s *Server) getArchiveListHandler(c *gin.Context) {
	ctx := context.Background()
	p := c.Query("page")
	ps := c.Query("size")

	if p == "" && ps == "" {
		// archives, err := archive.GetAll(ctx, s.repo, s.log)
		archives, err := s.repo.GetAllArchives(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, &models.ResponseError{
				Status:  "error",
				Message: err.Error(),
			})
			return
		}

		if len(archives) == 0 {
			c.JSON(http.StatusNotFound, &models.ResponseError{
				Status:  "error",
				Message: config.ErrNoArchive.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, archives)
	} else {

		page, _ := strconv.Atoi(p)
		if page == 0 {
			page = 1
		}
		pageSize, _ := strconv.Atoi(ps)
		if pageSize == 0 {
			pageSize = 10
		}

		archives, err := s.repo.GetArchiveList(ctx, repository.GetArchiveListParams{
			Limit:  int32(pageSize),
			Offset: (int32(page) - 1) * int32(pageSize),
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, &models.ResponseError{
				Status:  "error",
				Message: err.Error(),
			})
			return
		}
		if len(archives) == 0 {
			c.JSON(http.StatusNotFound, &models.ResponseError{
				Status:  "error",
				Message: config.ErrNoArchive.Error(),
			})
			return
		}
		count, err := s.repo.CountArchives(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, &models.ResponseError{
				Status:  "error",
				Message: err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"archives": archives,
			"total":    count,
		})
	}
}

func (s *Server) getArchiveHandler(c *gin.Context) {
	id := c.Param("id")

	ctx := context.Background()

	arch, err := s.repo.GetArchiveByID(ctx, id)
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
	tags, err := s.repo.GetArchiveTags(ctx, id)
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
	characters, err := s.repo.GetArchiveCharacters(ctx, id)
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
	parodies, err := s.repo.GetArchiveParodies(ctx, id)
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
	urls, err := s.repo.GetArchiveURLs(ctx, id)
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
	artists, err := s.repo.GetArchiveArtists(ctx, id)
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

	var pages int
	p, _ := filepath.Abs(*arch.ThumbsPath)
	d, err := os.ReadDir(filepath.Join(p, "pages"))
	if err != nil {
		pages = 0
	}
	for _, i := range d {
		if fsutil.MatchExtension(i.Name(), config.ImageExtensions) {
			pages++
		}
	}

	result := &models.ArchiveResponse{
		ArchiveID: id,
		Title:     arch.Title,
		Summary:   arch.Summary,
		Tags:      tags,
		Artist:    artists,
		Parody:    parodies,
		Character: characters,
		Language:  arch.Language,
		Category:  arch.Category,
		PageCount: arch.PageCount,
		Url:       urls,
		Hash:      arch.Hash,
		Pages:     pages,
		// ThumbsPath: archive.ThumbsPath,
		Type:        arch.Type,
		CreatedAt:   arch.CreatedAt,
		UpdatedAt:   arch.UpdatedAt,
		ReleaseDate: arch.ReleaseDate,
	}

	c.JSON(http.StatusOK, result)
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

	id := c.Param("id")

	ctx := context.Background()
	exists, err := s.repo.ArchiveIDExists(ctx, id)
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

	// TODO: Update handlers to use Metadata builder instead
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

	c.JSON(http.StatusOK, result)
}

// Delete

func (s *Server) deleteArchiveHandler(c *gin.Context) {
	id := c.Param("id")

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

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}
