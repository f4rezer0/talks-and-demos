package services

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"train-ticketing-ai/database"
	"train-ticketing-ai/models"
)

// GetAllStations retrieves all stations for autocomplete
func GetAllStations() ([]models.Station, error) {
	db := database.GetDB()
	rows, err := db.Query(`
		SELECT id, name, city, code
		FROM stations
		ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stations []models.Station
	for rows.Next() {
		var station models.Station
		err := rows.Scan(&station.ID, &station.Name, &station.City, &station.Code)
		if err != nil {
			return nil, err
		}
		stations = append(stations, station)
	}

	return stations, nil
}

// FindStationByNameOrCode finds a station by name or code (fuzzy match)
func FindStationByNameOrCode(query string) (*models.Station, error) {
	db := database.GetDB()

	// Try exact code match first
	var station models.Station
	err := db.QueryRow(`
		SELECT id, name, city, code
		FROM stations
		WHERE UPPER(code) = UPPER($1)
	`, query).Scan(&station.ID, &station.Name, &station.City, &station.Code)

	if err == nil {
		return &station, nil
	}

	// Try fuzzy match on name or city
	err = db.QueryRow(`
		SELECT id, name, city, code
		FROM stations
		WHERE UPPER(name) LIKE UPPER($1) OR UPPER(city) LIKE UPPER($1)
		ORDER BY similarity(name, $2) DESC
		LIMIT 1
	`, "%"+query+"%", query).Scan(&station.ID, &station.Name, &station.City, &station.Code)

	if err != nil {
		return nil, fmt.Errorf("station not found: %s", query)
	}

	return &station, nil
}

// SearchTrains searches for available trains based on criteria
func SearchTrains(req models.SearchRequest) ([]models.SearchResponse, error) {
	db := database.GetDB()

	// Find origin and destination stations
	origin, err := FindStationByNameOrCode(req.Origin)
	if err != nil {
		return nil, fmt.Errorf("origin station not found: %w", err)
	}

	destination, err := FindStationByNameOrCode(req.Destination)
	if err != nil {
		return nil, fmt.Errorf("destination station not found: %w", err)
	}

	// Parse date and get day of week
	travelDate, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}

	// Validate date is not in the past
	today := time.Now().Truncate(24 * time.Hour)
	if travelDate.Before(today) {
		return nil, fmt.Errorf("cannot book trains in the past")
	}

	// Validate date is not more than 90 days ahead
	maxDate := today.AddDate(0, 0, 90)
	if travelDate.After(maxDate) {
		return nil, fmt.Errorf("cannot book trains more than 90 days in advance")
	}

	dayOfWeek := int(travelDate.Weekday())

	if req.PassengerCount < 1 {
		req.PassengerCount = 1
	}

	// Build query with filters
	query := `
		SELECT
			s.id, s.train_id, s.origin_id, s.destination_id,
			s.departure_time, s.arrival_time, s.day_of_week,
			s.price_base, s.available_seats,
			t.id, t.number, t.type, t.has_wifi, t.has_food, t.total_seats,
			o.id, o.name, o.city, o.code,
			d.id, d.name, d.city, d.code
		FROM schedules s
		JOIN trains t ON s.train_id = t.id
		JOIN stations o ON s.origin_id = o.id
		JOIN stations d ON s.destination_id = d.id
		WHERE s.origin_id = $1
			AND s.destination_id = $2
			AND s.day_of_week = $3
			AND s.available_seats >= $4
	`

	args := []interface{}{origin.ID, destination.ID, dayOfWeek, req.PassengerCount}
	argIndex := 5

	// Time preference filter
	if req.TimePreference != "" && req.TimePreference != "any" {
		switch req.TimePreference {
		case "morning":
			query += fmt.Sprintf(" AND s.departure_time >= '06:00:00' AND s.departure_time < '12:00:00'")
		case "afternoon":
			query += fmt.Sprintf(" AND s.departure_time >= '12:00:00' AND s.departure_time < '18:00:00'")
		case "evening":
			query += fmt.Sprintf(" AND s.departure_time >= '18:00:00' AND s.departure_time < '24:00:00'")
		}
	}

	// WiFi filter
	if hasWifi, ok := req.Filters["has_wifi"].(bool); ok && hasWifi {
		query += " AND t.has_wifi = true"
	}

	// Food filter
	if hasFood, ok := req.Filters["has_food"].(bool); ok && hasFood {
		query += " AND t.has_food = true"
	}

	// Max price filter
	if maxPrice, ok := req.Filters["max_price"].(float64); ok && maxPrice > 0 {
		query += fmt.Sprintf(" AND s.price_base <= $%d", argIndex)
		args = append(args, maxPrice)
		argIndex++
	}

	query += " ORDER BY s.departure_time"

	log.Printf("Searching trains: origin=%s, destination=%s, date=%s, day_of_week=%d, passengers=%d",
		origin.Name, destination.Name, req.Date, dayOfWeek, req.PassengerCount)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error querying schedules: %w", err)
	}
	defer rows.Close()

	var results []models.SearchResponse
	for rows.Next() {
		var schedule models.Schedule
		var train models.Train
		var orig, dest models.Station

		err := rows.Scan(
			&schedule.ID, &schedule.TrainID, &schedule.OriginID, &schedule.DestinationID,
			&schedule.DepartureTime, &schedule.ArrivalTime, &schedule.DayOfWeek,
			&schedule.PriceBase, &schedule.AvailableSeats,
			&train.ID, &train.Number, &train.Type, &train.HasWifi, &train.HasFood, &train.TotalSeats,
			&orig.ID, &orig.Name, &orig.City, &orig.Code,
			&dest.ID, &dest.Name, &dest.City, &dest.Code,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning schedule: %w", err)
		}

		schedule.Train = train
		schedule.Origin = orig
		schedule.Destination = dest

		// Calculate duration
		duration := schedule.ArrivalTime.Sub(schedule.DepartureTime)
		hours := int(duration.Hours())
		minutes := int(duration.Minutes()) % 60

		// Format times for display
		departureStr := schedule.DepartureTime.Format("15:04")
		arrivalStr := schedule.ArrivalTime.Format("15:04")
		durationStr := fmt.Sprintf("%dh %dm", hours, minutes)

		// Calculate price (base price per person)
		pricePerPerson := schedule.PriceBase
		totalPrice := pricePerPerson * float64(req.PassengerCount)

		result := models.SearchResponse{
			Schedule:       schedule,
			DepartureTime:  departureStr,
			ArrivalTime:    arrivalStr,
			Duration:       durationStr,
			PricePerPerson: pricePerPerson,
			TotalPrice:     totalPrice,
		}

		results = append(results, result)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no trains found for the specified criteria")
	}

	return results, nil
}

// GetScheduleByID retrieves a schedule by ID
func GetScheduleByID(scheduleID int) (*models.Schedule, error) {
	db := database.GetDB()

	var schedule models.Schedule
	var train models.Train
	var origin, destination models.Station

	err := db.QueryRow(`
		SELECT
			s.id, s.train_id, s.origin_id, s.destination_id,
			s.departure_time, s.arrival_time, s.day_of_week,
			s.price_base, s.available_seats,
			t.id, t.number, t.type, t.has_wifi, t.has_food, t.total_seats,
			o.id, o.name, o.city, o.code,
			d.id, d.name, d.city, d.code
		FROM schedules s
		JOIN trains t ON s.train_id = t.id
		JOIN stations o ON s.origin_id = o.id
		JOIN stations d ON s.destination_id = d.id
		WHERE s.id = $1
	`, scheduleID).Scan(
		&schedule.ID, &schedule.TrainID, &schedule.OriginID, &schedule.DestinationID,
		&schedule.DepartureTime, &schedule.ArrivalTime, &schedule.DayOfWeek,
		&schedule.PriceBase, &schedule.AvailableSeats,
		&train.ID, &train.Number, &train.Type, &train.HasWifi, &train.HasFood, &train.TotalSeats,
		&origin.ID, &origin.Name, &origin.City, &origin.Code,
		&destination.ID, &destination.Name, &destination.City, &destination.Code,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("schedule not found")
		}
		return nil, err
	}

	schedule.Train = train
	schedule.Origin = origin
	schedule.Destination = destination

	return &schedule, nil
}

// CalculatePassengerPrice calculates price with discount based on passenger type
func CalculatePassengerPrice(basePrice float64, passengerType string) float64 {
	switch strings.ToLower(passengerType) {
	case "adult":
		return basePrice
	case "senior":
		return basePrice * 0.80 // 20% discount
	case "child":
		return basePrice * 0.70 // 30% discount
	case "infant":
		return 0.0 // Free, no seat
	default:
		return basePrice
	}
}
