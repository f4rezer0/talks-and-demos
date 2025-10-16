package models

// Station represents a train station
type Station struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	City string `json:"city"`
	Code string `json:"code"`
}
