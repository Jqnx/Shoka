package server

import (
	"context"
	"net/http"

	"Shoka/internal/fsutil"

	"github.com/gin-gonic/gin"
)

func (s *Server) testHandler(c *gin.Context) {
	ctx := context.Background()
	id := c.Param("id")

	arch, err := s.repo.GetArchiveByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
		return
	}

	zip, err := fsutil.OpenArchive(arch.FilePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
		return
	}
	list, err := zip.GetFileNames(false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
		return
	}

	c.JSON(http.StatusOK, len(list))
}
