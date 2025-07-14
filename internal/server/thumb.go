package server

import (
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
	"github.com/hibiken/asynq"
)

func (s *Server) generateThumbHandler(c *gin.Context) {
	id := c.Param("id")
	f := c.Query("force")
	var force bool

	if f != "" {
		b, err := strconv.ParseBool(f)
		if err != nil {
			c.JSON(http.StatusBadRequest, &models.ResponseError{
				Status:  "error",
				Message: "force has invalid boolean value",
			})
		}
		force = b
	}

	ctx := context.Background()

	arch, err := s.repo.GetArchiveByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.ResponseError{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	ch := make(chan *asynq.TaskInfo)

	if arch.ThumbsPath == nil {
		w := workers.NewWorkers(s.app, false, ctx)
		go w.Thumbs(ch, &arch)
		thumb := <-ch
		if thumb != nil {
			loc := fmt.Sprintf("/api/status/%s", thumb.ID)
			c.Header("Location", loc)
			c.JSON(http.StatusOK, &gin.H{
				"status":   "success",
				"location": loc,
			})
			return
		} else {
			c.JSON(http.StatusOK, &models.ResponseError{
				Status:  "success",
				Message: "set thumbs_path in database, thumbnails already existed on disk but archive's thumbs_path was missing from db",
			})
			return
		}
	}

	p, _ := filepath.Abs(*arch.ThumbsPath)
	if force {
		w := workers.NewWorkers(s.app, true, ctx)
		go w.Thumbs(ch, &arch)
		thumb := <-ch
		loc := fmt.Sprintf("/api/status/%s", thumb.ID)
		c.Header("Location", loc)
		c.JSON(http.StatusOK, &gin.H{
			"status":   "success",
			"location": loc,
		})
	} else {
		d, err := os.ReadDir(filepath.Join(p, "pages"))
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				w := workers.NewWorkers(s.app, false, ctx)
				go w.Thumbs(ch, &arch)
				thumb := <-ch
				loc := fmt.Sprintf("/api/status/%s", thumb.ID)
				c.Header("Location", loc)
				c.JSON(http.StatusOK, &gin.H{
					"status":   "success",
					"location": loc,
				})
			}
		} else if len(d) != int(arch.PageCount) {
			w := workers.NewWorkers(s.app, false, ctx)
			go w.Thumbs(ch, &arch)
			thumb := <-ch
			loc := fmt.Sprintf("/api/status/%s", thumb.ID)
			c.Header("Location", loc)
			c.JSON(http.StatusOK, &gin.H{
				"status":   "success",
				"location": loc,
			})
		} else {
			c.JSON(http.StatusBadRequest, &models.ResponseError{
				Status:  "failed",
				Message: "thumbnails already exist, try using force=true to force generate new ones",
			})
		}
	}
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
	_, err = os.Stat(pageDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			c.JSON(http.StatusInternalServerError, &models.ResponseError{
				Status:  "error",
				Message: "archives page directory does not exist",
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
