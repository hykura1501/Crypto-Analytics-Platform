package api

import (
	"database/sql"
	"net/http"

	"github.com/crypto-platform/crawler-service/internal/handler"
	"github.com/gin-gonic/gin"
)

func (s *Server) listSources(c *gin.Context) {
	sources, err := s.sourceHandler.ListSources()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"sources": sources,
		"total":   len(sources),
	})
}

func (s *Server) getSource(c *gin.Context) {
	sourceID := c.Param("id")
	source, err := s.sourceHandler.GetSource(sourceID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Source not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, source)
}

func (s *Server) createSource(c *gin.Context) {
	var req handler.CreateSourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.sourceHandler.CreateSource(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Trigger analysis in background
	go func() {
		if err := s.crawlService.AnalyzeSource(req.SourceID, req.RssURL); err != nil {
			// Log error but don't fail the request since source is created
			// In a real app, we might want to update source status
		}
	}()

	c.JSON(http.StatusCreated, gin.H{
		"message":   "Source created successfully",
		"source_id": req.SourceID,
	})
}

func (s *Server) updateSource(c *gin.Context) {
	sourceID := c.Param("id")
	var req handler.UpdateSourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.sourceHandler.UpdateSource(sourceID, req); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Source not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Source updated successfully",
		"source_id": sourceID,
	})
}

func (s *Server) deleteSource(c *gin.Context) {
	sourceID := c.Param("id")
	if err := s.sourceHandler.DeleteSource(sourceID); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Source not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Source deleted successfully",
		"source_id": sourceID,
	})
}
