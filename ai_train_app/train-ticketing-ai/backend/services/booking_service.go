package services

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"train-ticketing-ai/database"
	"train-ticketing-ai/models"
)

// CreateBooking creates a new booking with passengers
func CreateBooking(req models.BookingRequest) (*models.Booking, error) {
	db := database.GetDB()

	// Parse booking date
	bookingDate, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}

	// Validate date
	today := time.Now().Truncate(24 * time.Hour)
	if bookingDate.Before(today) {
		return nil, fmt.Errorf("cannot book trains in the past")
	}

	// Get schedule details
	schedule, err := GetScheduleByID(req.ScheduleID)
	if err != nil {
		return nil, fmt.Errorf("schedule not found: %w", err)
	}

	// Validate day of week matches
	if int(bookingDate.Weekday()) != schedule.DayOfWeek {
		return nil, fmt.Errorf("selected date does not match train schedule day")
	}

	// Count non-infant passengers (who need seats)
	seatsNeeded := 0
	for _, p := range req.Passengers {
		if p.PassengerType != "infant" {
			seatsNeeded++
		}
	}

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Check and lock available seats
	var availableSeats int
	err = tx.QueryRow(`
		SELECT available_seats
		FROM schedules
		WHERE id = $1
		FOR UPDATE
	`, req.ScheduleID).Scan(&availableSeats)

	if err != nil {
		return nil, fmt.Errorf("failed to check seat availability: %w", err)
	}

	if availableSeats < seatsNeeded {
		return nil, fmt.Errorf("insufficient seats available (need %d, have %d)", seatsNeeded, availableSeats)
	}

	// Generate booking reference
	bookingRef, err := generateBookingReference(tx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate booking reference: %w", err)
	}

	// Calculate total price and create passengers
	var totalPrice float64
	var passengers []models.Passenger
	seatNumber := 1

	for _, passengerReq := range req.Passengers {
		price := CalculatePassengerPrice(schedule.PriceBase, passengerReq.PassengerType)
		totalPrice += price

		passenger := models.Passenger{
			Name:          passengerReq.Name,
			PassengerType: passengerReq.PassengerType,
			Price:         price,
		}

		// Assign seat number (except for infants)
		if passengerReq.PassengerType != "infant" {
			passenger.SeatNumber = fmt.Sprintf("%dA", seatNumber)
			seatNumber++
		}

		passengers = append(passengers, passenger)
	}

	// Insert booking
	var bookingID int
	err = tx.QueryRow(`
		INSERT INTO bookings (booking_ref, schedule_id, booking_date, passenger_count, total_price, status)
		VALUES ($1, $2, $3, $4, $5, 'confirmed')
		RETURNING id
	`, bookingRef, req.ScheduleID, bookingDate, len(req.Passengers), totalPrice).Scan(&bookingID)

	if err != nil {
		return nil, fmt.Errorf("failed to create booking: %w", err)
	}

	// Insert passengers
	for i := range passengers {
		passengers[i].BookingID = bookingID

		err = tx.QueryRow(`
			INSERT INTO passengers (booking_id, name, passenger_type, seat_number, price)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id
		`, bookingID, passengers[i].Name, passengers[i].PassengerType,
			sql.NullString{String: passengers[i].SeatNumber, Valid: passengers[i].SeatNumber != ""},
			passengers[i].Price).Scan(&passengers[i].ID)

		if err != nil {
			return nil, fmt.Errorf("failed to add passenger: %w", err)
		}
	}

	// Update available seats
	_, err = tx.Exec(`
		UPDATE schedules
		SET available_seats = available_seats - $1
		WHERE id = $2
	`, seatsNeeded, req.ScheduleID)

	if err != nil {
		return nil, fmt.Errorf("failed to update seat availability: %w", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit booking: %w", err)
	}

	log.Printf("Booking created: %s for %d passengers on schedule %d", bookingRef, len(passengers), req.ScheduleID)

	// Create booking response
	booking := &models.Booking{
		ID:             bookingID,
		BookingRef:     bookingRef,
		ScheduleID:     req.ScheduleID,
		BookingDate:    bookingDate,
		PassengerCount: len(req.Passengers),
		TotalPrice:     totalPrice,
		Status:         "confirmed",
		CreatedAt:      time.Now(),
		Schedule:       *schedule,
		Passengers:     passengers,
	}

	return booking, nil
}

// GetBooking retrieves a booking by reference
func GetBooking(bookingRef string) (*models.Booking, error) {
	db := database.GetDB()

	var booking models.Booking
	var schedule models.Schedule
	var train models.Train
	var origin, destination models.Station

	// Get booking with schedule details
	err := db.QueryRow(`
		SELECT
			b.id, b.booking_ref, b.schedule_id, b.booking_date,
			b.passenger_count, b.total_price, b.status, b.created_at,
			s.id, s.train_id, s.origin_id, s.destination_id,
			s.departure_time, s.arrival_time, s.day_of_week,
			s.price_base, s.available_seats,
			t.id, t.number, t.type, t.has_wifi, t.has_food, t.total_seats,
			o.id, o.name, o.city, o.code,
			d.id, d.name, d.city, d.code
		FROM bookings b
		JOIN schedules s ON b.schedule_id = s.id
		JOIN trains t ON s.train_id = t.id
		JOIN stations o ON s.origin_id = o.id
		JOIN stations d ON s.destination_id = d.id
		WHERE b.booking_ref = $1
	`, bookingRef).Scan(
		&booking.ID, &booking.BookingRef, &booking.ScheduleID, &booking.BookingDate,
		&booking.PassengerCount, &booking.TotalPrice, &booking.Status, &booking.CreatedAt,
		&schedule.ID, &schedule.TrainID, &schedule.OriginID, &schedule.DestinationID,
		&schedule.DepartureTime, &schedule.ArrivalTime, &schedule.DayOfWeek,
		&schedule.PriceBase, &schedule.AvailableSeats,
		&train.ID, &train.Number, &train.Type, &train.HasWifi, &train.HasFood, &train.TotalSeats,
		&origin.ID, &origin.Name, &origin.City, &origin.Code,
		&destination.ID, &destination.Name, &destination.City, &destination.Code,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("booking not found")
		}
		return nil, err
	}

	schedule.Train = train
	schedule.Origin = origin
	schedule.Destination = destination
	booking.Schedule = schedule

	// Get passengers
	rows, err := db.Query(`
		SELECT id, booking_id, name, passenger_type, seat_number, price
		FROM passengers
		WHERE booking_id = $1
		ORDER BY id
	`, booking.ID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var passengers []models.Passenger
	for rows.Next() {
		var passenger models.Passenger
		var seatNumber sql.NullString

		err := rows.Scan(
			&passenger.ID, &passenger.BookingID, &passenger.Name,
			&passenger.PassengerType, &seatNumber, &passenger.Price,
		)
		if err != nil {
			return nil, err
		}

		if seatNumber.Valid {
			passenger.SeatNumber = seatNumber.String
		}

		passengers = append(passengers, passenger)
	}

	booking.Passengers = passengers

	return &booking, nil
}

// CancelBooking cancels a booking and restores seats
func CancelBooking(bookingRef string) error {
	db := database.GetDB()

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Get booking details
	var bookingID, scheduleID int
	var status string
	var seatsToRestore int

	err = tx.QueryRow(`
		SELECT b.id, b.schedule_id, b.status,
			(SELECT COUNT(*) FROM passengers WHERE booking_id = b.id AND passenger_type != 'infant')
		FROM bookings b
		WHERE b.booking_ref = $1
		FOR UPDATE
	`, bookingRef).Scan(&bookingID, &scheduleID, &status, &seatsToRestore)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("booking not found")
		}
		return err
	}

	if status == "cancelled" {
		return fmt.Errorf("booking already cancelled")
	}

	// Update booking status
	_, err = tx.Exec(`
		UPDATE bookings
		SET status = 'cancelled'
		WHERE id = $1
	`, bookingID)

	if err != nil {
		return fmt.Errorf("failed to cancel booking: %w", err)
	}

	// Restore seats
	_, err = tx.Exec(`
		UPDATE schedules
		SET available_seats = available_seats + $1
		WHERE id = $2
	`, seatsToRestore, scheduleID)

	if err != nil {
		return fmt.Errorf("failed to restore seats: %w", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit cancellation: %w", err)
	}

	log.Printf("Booking cancelled: %s (%d seats restored)", bookingRef, seatsToRestore)

	return nil
}

// generateBookingReference generates a unique booking reference
func generateBookingReference(tx *sql.Tx) (string, error) {
	year := time.Now().Year()

	// Get next sequence number for this year
	var count int
	err := tx.QueryRow(`
		SELECT COUNT(*) + 1
		FROM bookings
		WHERE EXTRACT(YEAR FROM created_at) = $1
	`, year).Scan(&count)

	if err != nil {
		return "", err
	}

	// Format: TRN-YYYY-NNNNN
	bookingRef := fmt.Sprintf("TRN-%d-%05d", year, count)

	// Check if it already exists (unlikely but possible with concurrent requests)
	var exists bool
	err = tx.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM bookings WHERE booking_ref = $1)
	`, bookingRef).Scan(&exists)

	if err != nil {
		return "", err
	}

	if exists {
		// Use timestamp-based fallback
		bookingRef = fmt.Sprintf("TRN-%d-%d", year, time.Now().UnixNano()%100000)
	}

	return bookingRef, nil
}
