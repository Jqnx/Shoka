package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Server) getAvailableSourcesHandler(c *gin.Context) {
	c.JSON(http.StatusOK, s.app.Sources.GetEnabledSources().Remote)
}

func (s *Server) getAllSources(c *gin.Context) {
	c.JSON(http.StatusOK, s.app.Sources)
}
