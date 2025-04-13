package server

import (
	"Shoka/cmd/web"
	"io/fs"
	"net/http"

	"github.com/a-h/templ"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := gin.Default()

	r.Use(gin.Recovery())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://*", "http://*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	staticFiles, _ := fs.Sub(web.Files, "assets")
	r.StaticFS("/assets", http.FS(staticFiles))
	r.GET("/web", func(c *gin.Context) {
		templ.Handler(web.HelloForm()).ServeHTTP(c.Writer, c.Request)
	})

	// Actual API
	// Archive API
	api := r.Group("/api")
	{
		archive := api.Group("/a")
		{
			archive.GET("/", s.getAllArchiveHandler)
			archive.GET("/:id", s.getArchiveHandler)
			archive.POST("/create", s.createArchiveHandler)
			archive.PUT("/:id", s.updateArchiveHandler)
			archive.DELETE("/:id", s.deleteArchiveHandler)
			// archive.GET("/lastid", s.getLastIDHandler)
		}

		// Artist API
		artist := api.Group("/artist")
		{
			artist.GET("/", s.getAllArtistHandler)
			artist.GET("/:name", s.getArtistHandler)
			artist.POST("/create", s.createArtistHandler)
			artist.PUT("/:name", s.updateArtistHandler)
			artist.DELETE("/:name", s.deleteArtistHandler)
		}

		// Group API
		group := api.Group("/group")
		{
			group.GET("/", s.getAllGroupHandler)
			group.GET("/:name", s.getGroupHandler)
			group.POST("/create", s.createGroupHandler)
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
		character := api.Group("/c")
		{
			character.GET("/:character", s.getArchiveByCharacterHandler)
			character.GET("/", s.getAllCharacterHandler)
		}

		// Parody API
		parody := api.Group("/p")
		{
			parody.GET("/:parody", s.getArchiveByParodyHandler)
			parody.GET("/", s.getAllParodyHandler)
		}
	}

	return r
}
