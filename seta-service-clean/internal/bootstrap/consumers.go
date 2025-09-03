package bootstrap

import (
	"context"
	"seta/internal/adapters/cache"
	"seta/internal/adapters/events"
	"seta/internal/infrastructure"

	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog"
)

// StartConsumers initializes and runs the Kafka consumers in the background.
func StartConsumers(ctx context.Context, log *zerolog.Logger, rdb *redis.Client) {
	// Initialize cache adapters that the consumer will use
	teamMemberCache := cache.NewRedisTeamCache(rdb)
	assetMetaCache := cache.NewRedisAssetMetaCache(rdb)
	aclCache := cache.NewRedisACLCache(rdb)

	// Initialize the event consumer adapter
	consumerAdapter := events.NewEventConsumer(log, teamMemberCache, assetMetaCache, aclCache)

	// Create readers for each topic
	teamActivityReader := infrastructure.NewKafkaReader("team.activity", "seta-service-group")
	assetChangesReader := infrastructure.NewKafkaReader("asset.changes", "seta-service-group")

	// Start each consumer in its own goroutine
	go func() {
		consumerAdapter.Start(ctx, teamActivityReader, consumerAdapter.HandleTeamEvent)
	}()

	go func() {
		consumerAdapter.Start(ctx, assetChangesReader, consumerAdapter.HandleAssetEvent)
	}()

	log.Info().Msg("Kafka consumers started")
}
