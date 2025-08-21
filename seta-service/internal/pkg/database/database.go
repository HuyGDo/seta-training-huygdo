package database

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Connect connects to the database and returns a GORM DB instance.
func Connect(log *zerolog.Logger) (*gorm.DB, error) {
	dsn := os.Getenv("DATABASE_URL")

	// close connection when shutdown application
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		// To enable sql query execution plan caching - need further testing for verification?
		PrepareStmt: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Info().Msg("Database connection successful.")
	return db, nil
}

var Rdb *redis.Client

func ConnectRedis(log *zerolog.Logger) error {
	redisAddr := os.Getenv("REDIS_URL")
	if redisAddr == "" {
		redisAddr = "localhost:6379" // Fallback for local dev
	}

	Rdb = redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := Rdb.Ping(ctx).Result(); err != nil {
		return fmt.Errorf("could not connect to Redis: %w", err)
	}

	log.Info().Msg("Redis connection successful.")
	return nil
}