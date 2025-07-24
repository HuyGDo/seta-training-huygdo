package database

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Connect connects to the database and returns a GORM DB instance.
func Connect() (*gorm.DB, error) {
	dsn := os.Getenv("DATABASE_URL")

	// close connection when shutdown application
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("Database connection successful.")
	return db, nil
}
