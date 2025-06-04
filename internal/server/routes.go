package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := gin.New()

	r.Use(gin.Recovery())
	r.Use(gin.Logger())
	//r.Use(cors.Default())
	//r.Use(cors.New(cors.Config{
	//	AllowOrigins:     []string{"*"},
	//	AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
	//	AllowHeaders:     []string{"Accept", "Authorization", "Content-Type", "Origin"},
	//	AllowCredentials: true,
	//}))

	r.StaticFS("/thumb", http.Dir("./thumb"))
	// Actual API
	// Archive API
	api := r.Group("/api")
	{
		archive := api.Group("/a")
		{
			archive.GET("/", s.getArchiveListHandler)
			archive.GET("/:id", s.getArchiveHandler)
			archive.GET("/:id/cover", s.getCoverHandler)
			archive.GET("/:id/:page", s.getThumbHandler)
			archive.GET("/:id/scanmeta", s.scanMetadataHandler)
			archive.POST("/", s.createArchiveHandler)
			archive.POST("/:id/cover", s.generateCoverHandler)
			archive.POST("/:id/thumb", s.generateThumbHandler)
			archive.PUT("/:id", s.updateArchiveHandler)
			// archive.DELETE("/:id", s.deleteArchiveHandler)
			// archive.GET("/lastid", s.getLastIDHandler)
		}

		// Artist API
		artist := api.Group("/artist")
		{
			artist.GET("/", s.getAllArtistHandler)
			artist.GET("/:name", s.getArtistHandler)
			artist.POST("/", s.createArtistHandler)
			artist.PUT("/:name", s.updateArtistHandler)
			artist.DELETE("/:name", s.deleteArtistHandler)
		}

		// Group API
		group := api.Group("/group")
		{
			group.GET("/", s.getAllGroupHandler)
			group.GET("/:name", s.getGroupHandler)
			group.POST("/", s.createGroupHandler)
			group.PUT("/:name", s.updateGroupHandler)
			group.DELETE("/:name", s.deleteGroupHandler)
		}

		// Tag API
		tag := api.Group("/tag")
		{
			tag.GET("/:tag", s.getArchiveByTagHandler)
			tag.GET("/", s.getAllTagHandler)
		}

		// Character API
		character := api.Group("/character")
		{
			character.GET("/:character", s.getArchiveByCharacterHandler)
			character.GET("/", s.getAllCharacterHandler)
		}

		// Parody API
		parody := api.Group("/parody")
		{
			parody.GET("/:parody", s.getArchiveByParodyHandler)
			parody.GET("/", s.getAllParodyHandler)
		}

		language := api.Group("/lang")
		{
			language.GET("/:language", s.getArchiveByLanguageHandler)
			language.GET("/", s.getAllLanguageHandler)
		}

	}

	return r
}
