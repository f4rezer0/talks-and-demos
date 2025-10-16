package models

// Train represents a train with its characteristics
type Train struct {
	ID         int    `json:"id"`
	Number     string `json:"number"`
	Type       string `json:"type"`
	HasWifi    bool   `json:"has_wifi"`
	HasFood    bool   `json:"has_food"`
	TotalSeats int    `json:"total_seats"`
}
