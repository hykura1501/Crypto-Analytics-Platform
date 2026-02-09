package api

import (
	"net/http"

	"github.com/crypto-platform/ai-service/internal/handler"
	"github.com/gin-gonic/gin"
)

type rssAnalysisRequest struct {
	XMLString string `json:"xml_string"`
}

func (s *Server) handleRss(c *gin.Context) {
	sourceId := c.Param("sourceId")
	var req rssAnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := s.rssHandler.Handle(c.Request.Context(), handler.RssMessage{
		SourceID:  sourceId,
		XMLString: req.XMLString,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "RSS analysis completed",
		"response": response,
	})
}
