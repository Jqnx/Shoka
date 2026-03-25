package server

import (
	"context"
	"errors"
	"net/http"

	"Shoka/internal/config"
	"Shoka/internal/database"
	"Shoka/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type FlaresolverrPayload struct {
	URL string `json:"url" binding:"url,required"`
}

func (s *Server) getFlaresolverrHandler(c *gin.Context) {
	conf := FlaresolverrPayload{
		URL: s.app.Cfg.Metadata.Flaresolverr.URL,
	}

	c.JSON(http.StatusOK, conf)
}

func (s *Server) setFlaresolverrHandler(c *gin.Context) {
	var req FlaresolverrPayload
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

	s.app.Cfg.Metadata.Flaresolverr.URL = req.URL

	viper.Set("sources.flaresolverr.url", req.URL)
	viper.WriteConfig()

	c.JSON(http.StatusOK, &models.Response{
		Status:  "success",
		Message: req,
	})
}

func (s *Server) resetDatabaseHandler(c *gin.Context) {
	ctx := context.Background()

	if err := database.ResetDatabase(ctx, s.app); err != nil {
		c.JSON(http.StatusInternalServerError, &models.Response{
			Status:  "error",
			Message: err,
		})
	}

	c.JSON(http.StatusOK, &models.Response{
		Status:  "success",
		Message: "successfully reset database",
	})
}
