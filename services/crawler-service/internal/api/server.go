package api

import (
	"database/sql"
	"net/http"

	"github.com/crypto-platform/crawler-service/config"
	"github.com/crypto-platform/crawler-service/internal/crawler"
	"github.com/crypto-platform/crawler-service/internal/handler"
	"github.com/gin-gonic/gin"
)

type Server struct {
	router         *gin.Engine
	config         *config.Config
	db             *sql.DB
	crawlService   *crawler.Service
	sourceHandler  *handler.SourceHandler
	articleHandler *handler.ArticleHandler
}

func NewServer(config *config.Config, db *sql.DB, crawlService *crawler.Service, sourceHandler *handler.SourceHandler) *Server {
	router := gin.Default()

	s := &Server{
		config:         config,
		db:             db,
		router:         router,
		crawlService:   crawlService,
		sourceHandler:  sourceHandler,
		articleHandler: handler.NewArticleHandler(db),
	}

	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	// API v1 Group
	v1 := s.router.Group("/api/v1/news")
	{
		// Infrastructure Health Check
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status":  "ok",
				"service": "crawler-service-go",
			})
		})

		// Crawler trigger
		v1.POST("/crawl/once", s.handleCrawlOnce)

		// Articles endpoint
		v1.GET("/articles", s.articleHandler.ListArticles)

		// Source CRUD endpoints
		v1.GET("/sources", s.listSources)
		v1.GET("/sources/:id", s.getSource)
		v1.POST("/sources", s.createSource)
		v1.PUT("/sources/:id", s.updateSource)
		v1.DELETE("/sources/:id", s.deleteSource)
		v1.POST("/sources/:id/analyze", s.analyzeSource)
	}
}

func (s *Server) Run() error {
	return s.router.Run(":" + s.config.Server.Port)
}

func (s *Server) handleCrawlOnce(c *gin.Context) {
	count, err := s.crawlService.CrawlOnce(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"saved": count,
	})
}
