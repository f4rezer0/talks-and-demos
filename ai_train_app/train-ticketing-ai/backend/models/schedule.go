package models

import "time"

// Schedule represents a train schedule with route and pricing information
type Schedule struct {
	ID             int       `json:"id"`
	TrainID        int       `json:"train_id"`
	OriginID       int       `json:"origin_id"`
	DestinationID  int       `json:"destination_id"`
	DepartureTime  time.Time `json:"departure_time"`
	ArrivalTime    time.Time `json:"arrival_time"`
	DayOfWeek      int       `json:"day_of_week"`
	PriceBase      float64   `json:"price_base"`
	AvailableSeats int       `json:"available_seats"`

	// Joined fields
	Train       Train   `json:"train"`
	Origin      Station `json:"origin"`
	Destination Station `json:"destination"`
}

// SearchRequest represents a train search query
type SearchRequest struct {
	Origin         string                 `json:"origin" binding:"required"`
	Destination    string                 `json:"destination" binding:"required"`
	Date           string                 `json:"date" binding:"required"`
	TimePreference string                 `json:"time_preference"` // morning, afternoon, evening, any
	PassengerCount int                    `json:"passenger_count"`
	Filters        map[string]interface{} `json:"filters"`
}

// SearchResponse represents a train search result with full details
type SearchResponse struct {
	Schedule       Schedule  `json:"schedule"`
	DepartureTime  string    `json:"departure_time"`
	ArrivalTime    string    `json:"arrival_time"`
	Duration       string    `json:"duration"`
	PricePerPerson float64   `json:"price_per_person"`
	TotalPrice     float64   `json:"total_price"`
}
