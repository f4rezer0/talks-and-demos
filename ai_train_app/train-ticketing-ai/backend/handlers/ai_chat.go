package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"train-ticketing-ai/models"
	"train-ticketing-ai/services"
)

// ChatWithAI processes AI chat messages
func ChatWithAI(c *gin.Context) {
	var req models.ChatRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("AI chat request - Session: %s, Message: %s", req.SessionID, req.Message)

	response, err := services.ProcessMessage(req.SessionID, req.Message)
	if err != nil {
		log.Printf("Error processing AI message: %v", err)
		c.JSON(http.StatusInternalServerError, models.ChatResponse{
			Success: false,
			Message: "I'm sorry, I encountered an error processing your request. Please try again.",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}
