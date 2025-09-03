package infrastructure

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectDB(log *zerolog.Logger) (*gorm.DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{PrepareStmt: true})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	log.Info().Msg("Database connection successful.")
	return db, nil
}

