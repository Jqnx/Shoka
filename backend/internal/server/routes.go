package server

import (
	"net/http"

	"Shoka/internal/config"
	"Shoka/internal/server/middleware"

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
		api.GET("/test", s.testHandler)
		api.GET("/search", s.searchArchiveHandler)
		archive := api.Group("/a")
		{
			// TODO: Merge getArchiveListHandler and getArchiveFilterHandler together into 1 handler
			archive.GET("", s.getArchiveListHandler)
			archive.POST("", s.createArchiveHandler)
			archive.GET("/recent", middleware.Auth(), s.getRecentlyReadHandler)
			archive.POST("/filter", s.getArchiveFilterHandler)
			archive.GET("/filters", s.getAllFiltersHandler)
			archive.POST("/shuffle", s.shuffleArchiveHandler)
			id := archive.Group("/:id")
			{
				id.GET("", s.getArchiveHandler)
				id.PUT("", s.updateArchiveHandler)
				id.GET("/cover", s.getCoverHandler)
				id.GET("/:page", s.getThumbHandler)
				id.POST("/:page", middleware.Auth(), s.updateReadingProgressHandler)
				id.POST("/cover", s.generateCoverHandler)
				id.POST("/thumb", s.generateThumbHandler)
				id.POST("/favorite", middleware.Auth(), s.favoriteArchiveHandler)
				id.DELETE("/rp", middleware.Auth(), s.deleteReadingProgressHandler)
				meta := id.Group("/meta")
				{
					meta.POST("/search", s.searchMetadataHandler)
					meta.POST("/scan", s.scanMetadataHandler)
					meta.POST("/tofile", s.metadataToFileHandler)
				}
			}
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
		/*
			group := api.Group("/group")
			{
				group.GET("", s.getAllGroupHandler)
				group.GET("/:name", s.getGroupHandler)
				group.POST("", s.createGroupHandler)
				group.PUT("/:name", s.updateGroupHandler)
				group.DELETE("/:name", s.deleteGroupHandler)
			}
		*/

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

		/*
			auth := api.Group("/auth")
			{
				auth.POST("/register", s.registerUser)
				auth.POST("/login", s.signInUser)
				auth.GET("/session", middleware.Auth(s.repo), s.getUserSession)
				auth.POST("/logout", middleware.Auth(s.repo), s.signOutUser)
			}
		*/

		user := api.Group("/user", middleware.Auth())
		{
			user.GET("/favorites", s.getUserFavoriteArchives)
			user.POST("/favorites", s.getFavoriteArchiveFilterHandler)
			// user.PUT("/update", middleware.Auth(s.repo), s.updateUserHandler)
		}

		/*
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
		*/
		config := api.Group("/config", middleware.Auth())
		{
			flare := config.Group("/flaresolverr")
			{
				flare.GET("", s.getFlaresolverrHandler)
				flare.POST("", s.setFlaresolverrHandler)

			}
			config.DELETE("/database", s.resetDatabaseHandler)
		}
	}

	return r
}
