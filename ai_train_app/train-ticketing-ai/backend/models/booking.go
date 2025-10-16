package models

import "time"

// Booking represents a train ticket booking
type Booking struct {
	ID             int        `json:"id"`
	BookingRef     string     `json:"booking_ref"`
	ScheduleID     int        `json:"schedule_id"`
	BookingDate    time.Time  `json:"booking_date"`
	PassengerCount int        `json:"passenger_count"`
	TotalPrice     float64    `json:"total_price"`
	Status         string     `json:"status"`
	CreatedAt      time.Time  `json:"created_at"`

	// Joined fields
	Schedule   Schedule    `json:"schedule"`
	Passengers []Passenger `json:"passengers"`
}

// Passenger represents a passenger in a booking
type Passenger struct {
	ID            int     `json:"id"`
	BookingID     int     `json:"booking_id"`
	Name          string  `json:"name"`
	PassengerType string  `json:"passenger_type"` // adult, senior, child, infant
	SeatNumber    string  `json:"seat_number"`
	Price         float64 `json:"price"`
}

// BookingRequest represents a booking creation request
type BookingRequest struct {
	ScheduleID int                      `json:"schedule_id" binding:"required"`
	Date       string                   `json:"date" binding:"required"`
	Passengers []PassengerCreateRequest `json:"passengers" binding:"required,min=1"`
}

// PassengerCreateRequest represents passenger data for booking
type PassengerCreateRequest struct {
	Name          string `json:"name" binding:"required"`
	PassengerType string `json:"passenger_type" binding:"required"`
}

// BookingResponse represents a booking creation response
type BookingResponse struct {
	Success bool     `json:"success"`
	Message string   `json:"message"`
	Booking *Booking `json:"booking,omitempty"`
}
