package server

import (
	"Shoka/internal/config"
	"Shoka/internal/server/middleware"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := gin.New()

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("useragent", config.ValidateUserAgent)
	}

	r.Use(gin.Recovery())
	r.Use(gin.Logger())
	r.Use(cors.Default())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type", "Origin"},
		AllowCredentials: true,
	}))

	// Static Assets
	r.Static("/assets", "./assets")

	// Websocket
	r.GET("/ws", s.app.Hub.HandleWebSocket)

	// Actual API
	// Archive API
	api := r.Group("/api")
	{
		api.POST("/test", s.testHandler)
		api.GET("/search", s.searchArchiveHandler)
		archive := api.Group("/a")
		{
			archive.GET("", s.getArchiveListHandler)
			archive.GET("/recent", middleware.Auth(s.repo), s.getRecentlyReadHandler)
			archive.POST("/filter", s.getArchiveFilterHandler)
			archive.GET("/filters", s.getAllFiltersHandler)
			archive.GET("/:id", s.getArchiveHandler)
			archive.GET("/:id/cover", s.getCoverHandler)
			archive.GET("/:id/:page", s.getThumbHandler)
			archive.GET("/:id/scanmeta", s.scanMetadataHandler)
			archive.POST("/:id/search", s.searchMetadataHandler)
			archive.POST("/shuffle", s.shuffleArchiveHandler)
			archive.POST("", s.createArchiveHandler)
			archive.POST("/:id/cover", s.generateCoverHandler)
			archive.POST("/:id/thumb", s.generateThumbHandler)
			archive.POST("/:id/favorite", middleware.Auth(s.repo), s.favoriteArchiveHandler)
			archive.POST("/:id/:page", middleware.Auth(s.repo), s.updateReadingProgressHandler)
			archive.PUT("/:id", s.updateArchiveHandler)
			archive.DELETE("/:id/rp", middleware.Auth(s.repo), s.deleteReadingProgressHandler)
			// archive.DELETE("/:id", s.deleteArchiveHandler)
			// archive.GET("/lastid", s.getLastIDHandler)
		}

		// Artist API
		artist := api.Group("/artist")
		{
			artist.GET("", s.getAllArtistHandler)
			artist.GET("/:name", s.getArtistHandler)
			artist.POST("", s.createArtistHandler)
			artist.PUT("/:name", s.updateArtistHandler)
			artist.DELETE("/:name", s.deleteArtistHandler)
		}

		// Group API
		group := api.Group("/group")
		{
			group.GET("", s.getAllGroupHandler)
			group.GET("/:name", s.getGroupHandler)
			group.POST("", s.createGroupHandler)
			group.PUT("/:name", s.updateGroupHandler)
			group.DELETE("/:name", s.deleteGroupHandler)
		}

		// Tag API
		tag := api.Group("/tag")
		{
			tag.GET("/:tag", s.getArchiveByTagHandler)
			tag.GET("", s.getAllTagHandler)
		}

		// Character API
		character := api.Group("/character")
		{
			character.GET("/:character", s.getArchiveByCharacterHandler)
			character.GET("", s.getAllCharacterHandler)
		}

		// Parody API
		parody := api.Group("/parody")
		{
			parody.GET("/:parody", s.getArchiveByParodyHandler)
			parody.GET("", s.getAllParodyHandler)
		}

		language := api.Group("/lang")
		{
			language.GET("/:language", s.getArchiveByLanguageHandler)
			language.GET("", s.getAllLanguageHandler)
		}

		category := api.Group("/category")
		{
			category.GET("/:category", s.getArchiveByCategoryHandler)
			category.GET("", s.getAllCategoryHandler)
		}

		auth := api.Group("/auth")
		{
			auth.POST("/register", s.registerUser)
			auth.POST("/login", s.signInUser)
			auth.GET("/session", middleware.Auth(s.repo), s.getUserSession)
			auth.POST("/logout", middleware.Auth(s.repo), s.signOutUser)
		}

		user := api.Group("/user")
		{
			// user.GET("/favorites", middleware.Auth(s.repo), s.getUserFavoriteArchives)
			user.POST("/favorites", middleware.Auth(s.repo), s.getFavoriteArchiveFilterHandler)
			user.PUT("/update", middleware.Auth(s.repo), s.updateUserHandler)
			user.DELETE("/delete", middleware.Auth(s.repo), s.deleteUserHandler)
		}

		download := api.Group("/download")
		{
			download.POST("", s.addDownloadHandler)
			download.GET("", s.getAllDownloadsHandler)
			download.GET("/:id", s.getDownloadHandler)
			download.GET("/active", s.getActiveDownloadHandler)
			download.POST("/:id/pause", s.pauseDownloadHandler)
			download.POST("/:id/resume", s.resumeDownloadHandler)
			download.DELETE("/:id", s.deleteDownloadHandler)
		}
		config := api.Group("/config")
		{
			config.GET("/nh", s.getNHCredentialsHandler)
			config.POST("/nh", s.setNHCredentialsHandler)
		}
	}

	return r
}
