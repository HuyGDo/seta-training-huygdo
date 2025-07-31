package database

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Connect connects to the database and returns a GORM DB instance.
func Connect(log *logrus.Logger) (*gorm.DB, error) {
	dsn := os.Getenv("DATABASE_URL")

	// close connection when shutdown application
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Info("Database connection successful.")
	return db, nil
}
