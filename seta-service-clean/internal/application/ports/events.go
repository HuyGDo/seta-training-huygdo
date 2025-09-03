package ports

import (
	"context"
	"time"
)

// EventPayload is a generic structure for events published by the application.
type EventPayload struct {
	EventType    string
	TeamID       string
	AssetType    string
	AssetID      string
	OwnerID      string
	ActionBy     string
	TargetUserID string
	Timestamp    time.Time
}

// EventPublisher defines the interface for publishing domain events.
type EventPublisher interface {
	PublishTeamEvent(ctx context.Context, payload EventPayload) error
	PublishAssetEvent(ctx context.Context, payload EventPayload) error
}

// Outbox defines an interface for transactional outbox pattern.
// This ensures that events are published only if the database transaction succeeds.
type Outbox interface {
	Save(ctx context.Context, payload EventPayload) error
}
