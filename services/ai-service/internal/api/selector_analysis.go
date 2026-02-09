package api

import (
	"net/http"

	"github.com/crypto-platform/ai-service/internal/handler"
	"github.com/gin-gonic/gin"
)

type selectorAnalysisRequest struct {
	HTMLString string `json:"html_string"`
}

func (s *Server) handleSelector(c *gin.Context) {
	sourceId := c.Param("sourceId")
	var req selectorAnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := s.selectorHandler.Handle(c.Request.Context(), handler.SelectorMessage{
		SourceID:   sourceId,
		HTMLString: req.HTMLString,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "CSS selector analysis completed",
		"response": response,
	})
}
