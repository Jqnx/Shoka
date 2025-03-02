package server

import (
	"Shoka/internal/store"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CreateTagPayload struct {
	Name string `json:"name"`
}

func (s *Server) createTagHandler(c *gin.Context) {
	var payload CreateTagPayload

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"bad request": err.Error()})
		return
	}

	fmt.Println(payload.Name)

	tag := &store.Tag{
		Name: payload.Name,
	}

	if err := s.store.Tags.Create(c, tag); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"internal error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"status": "tag created"})
}
