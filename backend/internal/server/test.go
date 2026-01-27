package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Server) testHandler(c *gin.Context) {
	c.JSON(http.StatusOK, &gin.H{
		"test": len(""),
	})
}
