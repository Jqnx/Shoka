package server

import (
	"io/fs"
	"net/http"

	"Shoka/cmd/web"

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

	r.GET("/", s.HelloWorldHandler)

	staticFiles, _ := fs.Sub(web.Files, "assets")
	r.StaticFS("/assets", http.FS(staticFiles))
	r.GET("/web", func(c *gin.Context) {
		templ.Handler(web.HelloForm()).ServeHTTP(c.Writer, c.Request)
	})

	r.POST("/hello", func(c *gin.Context) {
		web.HelloWebHandler(c.Writer, c.Request)
	})

	// Actual API
	// Archive API
	archive := r.Group("/a")
	{
		archive.GET("/", s.getAllArchiveHandler)
		archive.GET("/:id", s.getArchiveHandler)
		archive.POST("/create", s.createArchiveHandler)
		archive.PUT("/:id", s.updateArchiveHandler)
		archive.DELETE("/:id", s.deleteArchiveHandler)
		archive.GET("/lastid", s.getLastIDHandler)
	}

	// Artist API
	artist := r.Group("/artist")
	{
		artist.GET("/", s.getAllArtistHandler)
		artist.GET("/:name", s.getArtistHandler)
		artist.POST("/create", s.createArtistHandler)
	}

	return r
}

func (s *Server) HelloWorldHandler(c *gin.Context) {
	resp := make(map[string]string)
	resp["message"] = "Hello World"

	c.JSON(http.StatusOK, resp)
}
