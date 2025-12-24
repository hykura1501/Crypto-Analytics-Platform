package api

import (
	"database/sql"

	"github.com/crypto-platform/ai-service/config"
	"github.com/crypto-platform/ai-service/internal/handler"
	"github.com/gin-gonic/gin"
)

type Server struct {
	router     *gin.Engine
	config     *config.Config
	db         *sql.DB
	rssHandler *handler.RssHandler
}

func NewServer(config *config.Config, db *sql.DB, rssHandler *handler.RssHandler) *Server {
	router := gin.New()

	s := &Server{
		config:     config,
		db:         db,
		router:     router,
		rssHandler: rssHandler,
	}

	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	// API Group
	v1 := s.router.Group("/api/v1/ai")
	{
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"service": "ai-service", "status": "running"})
		})
		// Add more routes here
		v1.POST("/rss/:sourceId", s.handleRss)
	}
}

func (s *Server) Run() error {
	return s.router.Run(":" + s.config.API.Port)
}
