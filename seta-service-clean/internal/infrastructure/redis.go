package infrastructure

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog"
)

func ConnectRedis(log *zerolog.Logger) (*redis.Client, error) {
	redisAddr := os.Getenv("REDIS_URL")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := rdb.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("could not connect to Redis: %w", err)
	}
	log.Info().Msg("Redis connection successful.")
	return rdb, nil
}