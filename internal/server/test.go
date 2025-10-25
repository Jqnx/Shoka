package server

import (
	"Shoka/internal/language"
	"Shoka/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Server) testHandler(c *gin.Context) {
	lang := "en"

	conv := language.NewLanguageConverter()

	iso, err := conv.ToISO(lang)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.Response{
			Status:  "error",
			Message: err.Error(),
		})
	}

	c.JSON(http.StatusOK, iso)
}
