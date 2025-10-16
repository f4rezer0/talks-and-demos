package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"train-ticketing-ai/models"
	"train-ticketing-ai/services"
)

// CreateBooking creates a new booking
func CreateBooking(c *gin.Context) {
	var req models.BookingRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Booking request: %+v", req)

	booking, err := services.CreateBooking(req)
	if err != nil {
		log.Printf("Error creating booking: %v", err)
		c.JSON(http.StatusBadRequest, models.BookingResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.BookingResponse{
		Success: true,
		Message: "Booking created successfully",
		Booking: booking,
	})
}

// GetBooking retrieves a booking by reference
func GetBooking(c *gin.Context) {
	bookingRef := c.Param("ref")

	booking, err := services.GetBooking(bookingRef)
	if err != nil {
		log.Printf("Error getting booking: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
		return
	}

	c.JSON(http.StatusOK, booking)
}

// CancelBooking cancels a booking
func CancelBooking(c *gin.Context) {
	bookingRef := c.Param("ref")

	err := services.CancelBooking(bookingRef)
	if err != nil {
		log.Printf("Error cancelling booking: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("Booking %s cancelled successfully", bookingRef),
	})
}
