package database

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Connect connects to the database and returns a GORM DB instance.
func Connect(log *zerolog.Logger) (*gorm.DB, error) {
	// put to config package
	// manage environment in one place so can easily check, validate, and change
	dsn := os.Getenv("DATABASE_URL")

	// close connection when shutdown application
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Info().Msg("Database connection successful.")
	return db, nil
}
