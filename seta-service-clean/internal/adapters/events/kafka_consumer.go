package events

import (
	"context"
	"encoding/json"
	"seta/internal/application/ports"
	"time"

	"github.com/rs/zerolog"
	"github.com/segmentio/kafka-go"
)

// EventPayload defines the structure of the JSON messages from Kafka.
type EventPayload struct {
	EventType    string    `json:"eventType"`
	TeamID       string    `json:"teamId,omitempty"`
	AssetType    string    `json:"assetType,omitempty"`
	AssetID      string    `json:"assetId,omitempty"`
	OwnerID      string    `json:"ownerId,omitempty"`
	ActionBy     string    `json:"actionBy"`
	TargetUserID string    `json:"targetUserId,omitempty"`
	Timestamp    time.Time `json:"timestamp"`
}

// EventConsumer handles incoming Kafka events and updates the cache.
type EventConsumer struct {
	log             *zerolog.Logger
	teamMemberCache ports.TeamMemberCache
	assetMetaCache  ports.AssetMetaCache
	aclCache        ports.ACLCache
}

// NewEventConsumer creates a new EventConsumer.
func NewEventConsumer(log *zerolog.Logger, tmc ports.TeamMemberCache, amc ports.AssetMetaCache, ac ports.ACLCache) *EventConsumer {
	return &EventConsumer{
		log:             log,
		teamMemberCache: tmc,
		assetMetaCache:  amc,
		aclCache:        ac,
	}
}

// Start consuming messages from a given Kafka reader.
func (c *EventConsumer) Start(ctx context.Context, reader *kafka.Reader, handler func(context.Context, kafka.Message)) {
	c.log.Info().Str("topic", reader.Config().Topic).Msg("Starting Kafka consumer")
	for {
		select {
		case <-ctx.Done():
			c.log.Info().Str("topic", reader.Config().Topic).Msg("Stopping Kafka consumer")
			reader.Close()
			return
		default:
			msg, err := reader.ReadMessage(ctx)
			if err != nil {
				c.log.Error().Err(err).Str("topic", reader.Config().Topic).Msg("Error reading Kafka message")
				continue
			}
			handler(ctx, msg)
		}
	}
}

// HandleTeamEvent processes messages from the "team.activity" topic.
func (c *EventConsumer) HandleTeamEvent(ctx context.Context, m kafka.Message) {
	var payload EventPayload
	if err := json.Unmarshal(m.Value, &payload); err != nil {
		c.log.Error().Err(err).Msg("Error unmarshalling team event")
		return
	}

	c.log.Info().Str("event_type", payload.EventType).Str("team_id", payload.TeamID).Msg("Processing team event")

	switch payload.EventType {
	case "MEMBER_ADDED":
		c.teamMemberCache.AddTeamMember(ctx, payload.TeamID, payload.TargetUserID)
	case "MEMBER_REMOVED":
		c.teamMemberCache.RemoveTeamMember(ctx, payload.TeamID, payload.TargetUserID)
	}
}

// HandleAssetEvent processes messages from the "asset.changes" topic.
func (c *EventConsumer) HandleAssetEvent(ctx context.Context, m kafka.Message) {
	var payload EventPayload
	if err := json.Unmarshal(m.Value, &payload); err != nil {
		c.log.Error().Err(err).Msg("Error unmarshalling asset event")
		return
	}

	c.log.Info().Str("event_type", payload.EventType).Str("asset_id", payload.AssetID).Msg("Processing asset event")

	// Always invalidate the main asset cache on any change.
	if payload.AssetType != "" && payload.AssetID != "" {
		c.assetMetaCache.InvalidateAsset(ctx, payload.AssetType, payload.AssetID)
	}

	// Additionally, invalidate the ACL cache when sharing permissions change.
	switch payload.EventType {
	case "FOLDER_SHARED", "NOTE_SHARED", "FOLDER_UNSHARED", "NOTE_UNSHARED":
		c.aclCache.InvalidateACL(ctx, payload.AssetID)
	}
}

