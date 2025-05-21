package server

import (
	"Shoka/internal/config"
	"Shoka/internal/fsutil"
	"Shoka/internal/models"
	"Shoka/internal/workers"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

func (s *Server) generateThumbHandler(c *gin.Context) {
	id := c.Param("id")
	// force := c.Query("force")

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
	p, _ := filepath.Abs(*arch.ThumbsPath)

	d, err := os.ReadDir(filepath.Join(p, "pages"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			w := workers.NewWorkers(s.app, false, ctx)
			go w.Thumbs(&arch)
		}
	} else if len(d) < int(arch.PageCount) {
		w := workers.NewWorkers(s.app, false, ctx)
		go w.Thumbs(&arch)
	}
	c.JSON(http.StatusOK, &gin.H{
		"status": "success",
	})
}

func (s *Server) getThumbHandler(c *gin.Context) {
	id := c.Param("id")
	page, _ := strconv.Atoi(c.Param("page"))

	ctx := context.Background()

	arch, err := s.repo.GetArchiveByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.ResponseError{
			Status:  "error",
			Message: err,
		})
		return
	}

	if arch.ThumbsPath == nil {
		c.JSON(http.StatusInternalServerError, &models.ResponseError{
			Status:  "error",
			Message: "archive has no thumbspath",
		})
		return
	}

	if *arch.ThumbsPath == "" {
		c.JSON(http.StatusInternalServerError, &models.ResponseError{
			Status:  "error",
			Message: "archive has no thumbspath",
		})
		return
	}

	pageDir := filepath.Join(*arch.ThumbsPath, "pages")
	var file string

	cont := fsutil.ArchiveContents(*arch.FilePath)
	pageToIndex := page - 1
	for i, f := range cont {
		if i == pageToIndex {
			n := fsutil.StripExtension(f)
			file = fmt.Sprintf("%s.webp", n)
		}
	}

	full := filepath.Join(pageDir, file)
	if !fsutil.FileExists(full) {
		w, err := fsutil.WatchFile(pageDir, full)
		if err != nil {
			log.Println("error:", err)
		}
		c.File(w)
	} else {
		c.File(full)
	}
}
