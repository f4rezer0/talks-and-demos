package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"train-ticketing-ai/models"
	"train-ticketing-ai/services"
)

// GetStations returns all available stations
func GetStations(c *gin.Context) {
	stations, err := services.GetAllStations()
	if err != nil {
		log.Printf("Error getting stations: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve stations"})
		return
	}

	c.JSON(http.StatusOK, stations)
}

// SearchTrains searches for available trains
func SearchTrains(c *gin.Context) {
	var req models.SearchRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Search request: %+v", req)

	results, err := services.SearchTrains(req)
	if err != nil {
		log.Printf("Error searching trains: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, results)
}

// GetTrain returns train details by schedule ID
func GetTrain(c *gin.Context) {
	scheduleID := c.Param("id")

	var id int
	if _, err := fmt.Sscanf(scheduleID, "%d", &id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid schedule ID"})
		return
	}

	schedule, err := services.GetScheduleByID(id)
	if err != nil {
		log.Printf("Error getting train: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Train not found"})
		return
	}

	c.JSON(http.StatusOK, schedule)
}
