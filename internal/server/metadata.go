package server

import (
	"context"
	"errors"
	"net/http"
	"slices"

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
	ctx := context.Background()
	id := c.Param("id")

	// Get Metadata from Database
	// TODO: Create function for getting archive + all metadata
	arch, err := s.repo.GetArchiveByID(ctx, id)
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
	tags, err := s.repo.GetArchiveTags(ctx, id)
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
	characters, err := s.repo.GetArchiveCharacters(ctx, id)
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
	parodies, err := s.repo.GetArchiveParodies(ctx, id)
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
	urls, err := s.repo.GetArchiveURLs(ctx, id)
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
	artists, err := s.repo.GetArchiveArtists(ctx, id)
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
	ci.SetMetadata(comicinfo.ComicInfoParams{
		Archive:   arch,
		Artists:   artists,
		Tags:      tags,
		Parody:    parodies,
		Character: characters,
		URLs:      urls,
	})

	if err := ci.Write(arch.FilePath, s.app.Cfg.TempDir); err != nil {
		c.JSON(http.StatusInternalServerError, &models.Response{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	// TODO: Update file hash on ci.Write, make it a function in the archive package that also moves thumbs, cover and pages to new folder for correct hash
	// TODO: Probably run this as a job or in a goroutine (asyncronous)

	c.JSON(http.StatusCreated, &models.Response{
		Status:  "success",
		Message: "comicinfo file created",
	})
}
