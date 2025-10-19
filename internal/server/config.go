package server

import (
	"errors"
	"net/http"

	"Shoka/internal/config"
	"Shoka/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type NHLoginPayload struct {
	Csrftoken string `json:"csrftoken" binding:"required"`
	Useragent string `json:"useragent" binding:"useragent,required"`
}

type FlaresolverrPayload struct {
	URL string `json:"url" binding:"url,required"`
}

func (s *Server) getNHCredentialsHandler(c *gin.Context) {
	creds := NHLoginPayload{
		Csrftoken: s.app.Cfg.Sources.NHentai.CSRFToken,
		Useragent: s.app.Cfg.Sources.NHentai.UserAgent,
	}

	c.JSON(http.StatusOK, creds)
}

func (s *Server) setNHCredentialsHandler(c *gin.Context) {
	var req NHLoginPayload
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

	s.app.Cfg.Sources.NHentai.CSRFToken = req.Csrftoken
	s.app.Cfg.Sources.NHentai.UserAgent = req.Useragent

	viper.Set("sources.nhentai.csrftoken", req.Csrftoken)
	viper.Set("sources.nhentai.useragent", req.Useragent)
	viper.WriteConfig()

	c.JSON(http.StatusOK, &models.Response{
		Status:  "success",
		Message: req,
	})
}

func (s *Server) getFlaresolverrHandler(c *gin.Context) {
	conf := FlaresolverrPayload{
		URL: s.app.Cfg.Sources.Flaresolverr.URL,
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

	s.app.Cfg.Sources.Flaresolverr.URL = req.URL

	viper.Set("sources.flaresolverr.url", req.URL)
	viper.WriteConfig()

	c.JSON(http.StatusOK, &models.Response{
		Status:  "success",
		Message: req,
	})
}
