package server

import (
	"Shoka/internal/models"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CreateArtistPayload struct {
	Name  string         `json:"name" binding:"required"`
	Alias []models.Alias `json:"aliases"`
	Group []models.Group `json:"groups"`
	Links []models.Links `json:"links"`
}

func (s *Server) createArtistHandler(c *gin.Context) {
	var payload CreateArtistPayload

	if errPost := c.ShouldBind(&payload); errPost != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errPost.Error()})
		return
	}

	artist := &models.Artist{
		Name:  payload.Name,
		Alias: payload.Alias,
		Group: payload.Group,
		Links: payload.Links,
	}

	fmt.Println(payload.Name)
	fmt.Println(payload.Alias)
	fmt.Println(payload.Links)

	if err := s.store.Artist.Create(c, artist); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"internal error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"created": artist})
}
