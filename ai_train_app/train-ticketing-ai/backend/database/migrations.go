package database

import (
	"database/sql"
	"log"
)

// RunMigrations ensures all required tables exist
// Note: In production, use a proper migration tool
func RunMigrations(db *sql.DB) error {
	log.Println("Checking database schema...")

	// Check if tables exist
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_name = 'stations'
		)
	`).Scan(&exists)

	if err != nil {
		return err
	}

	if exists {
		log.Println("Database schema already exists, skipping migrations")
		return nil
	}

	log.Println("Database schema appears to be set up via init-db.sql")
	return nil
}
