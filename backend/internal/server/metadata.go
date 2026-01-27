package server

import (
	"context"
	"errors"
	"net/http"
	"slices"

	"Shoka/internal/archive"
	"Shoka/internal/config"
	"Shoka/internal/models"
	"Shoka/internal/sources"
	"Shoka/internal/sources/comicinfo"
	"Shoka/internal/workers"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"
)

type MetadataRequest struct {
	Title  string `json:"title" binding:"required"`
	Source string `json:"source" binding:"required"`
}

func (s *Server) scanMetadataHandler(c *gin.Context) {
	ctx := context.Background()
	id := c.Param("id")
	src := c.Query("source")

	archive, err := s.repo.GetArchiveByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, &models.Response{
				Status:  "error",
				Message: config.ErrArchiveNotFound.Error(),
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

	w := workers.NewWorkers(s.app, false, ctx)
	go w.Metadata(&archive, src)
	c.JSON(http.StatusOK, "success")
}

func (s *Server) searchMetadataHandler(c *gin.Context) {
	var req MetadataRequest
	if errPost := c.ShouldBind(&req); errPost != nil {
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

	if ok := slices.Contains(config.SourcesList, req.Source); !ok {
		c.JSON(http.StatusBadRequest, &models.ResponseFail{
			Status: "failed",
			Data:   "invalid source",
		})
		return
	}

	source, err := sources.NewSource(s.app.Cfg, req.Source, &config.MethodTitle)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.Response{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}
	source.SetTitle(req.Title)

	meta, err := source.GetMetadata()
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.Response{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, meta)
}

func (s *Server) metadataToFileHandler(c *gin.Context) {
	id := c.Param("id")

	// Get Metadata from Database
	arch, err := archive.GetWithMetadata(s.app, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, &models.Response{
				Status:  "error",
				Message: config.ErrArchiveNotFound.Error(),
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

	ci := comicinfo.NewComicInfo()
	if err := ci.SetMetadata(comicinfo.ComicInfoParams{
		Archive:   arch.Archive,
		Artists:   arch.Artists,
		Tags:      arch.Tags,
		Parody:    arch.Parodies,
		Character: arch.Characters,
		URLs:      arch.URLs,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, &models.Response{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	if err := ci.Write(arch.Archive.FilePath, s.app.Cfg.TempDir); err != nil {
		c.JSON(http.StatusInternalServerError, &models.Response{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	a := archive.RepoToArchive(arch.Archive, s.app)
	if err := a.MoveOnFileUpdate(); err != nil {
		c.JSON(http.StatusInternalServerError, &models.Response{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, &models.Response{
		Status:  "success",
		Message: "comicinfo file created",
	})
}
