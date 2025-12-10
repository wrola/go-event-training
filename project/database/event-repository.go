package database

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type EventRepository struct {
	db *sqlx.DB
}

func NewEventRepository(db *sqlx.DB) EventRepository {
	if db == nil {
		panic("db is nil")
	}
	return EventRepository{db: db}
}

type StoredEvent struct {
	EventID      uuid.UUID       `db:"event_id"`
	PublishedAt  time.Time       `db:"published_at"`
	EventName    string          `db:"event_name"`
	EventPayload json.RawMessage `db:"event_payload"`
}

func (r EventRepository) StoreEvent(
	ctx context.Context,
	eventID string,
	publishedAt time.Time,
	eventName string,
	eventPayload []byte,
) error {
	// Validate eventID is valid UUID
	parsedEventID, err := uuid.Parse(eventID)
	if err != nil {
		return fmt.Errorf("invalid event_id: %w", err)
	}

	// Convert raw payload bytes to json.RawMessage for JSONB storage
	jsonPayload := json.RawMessage(eventPayload)

	_, err = r.db.ExecContext(
		ctx,
		`INSERT INTO events (event_id, published_at, event_name, event_payload)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (event_id) DO NOTHING`,
		parsedEventID,
		publishedAt,
		eventName,
		jsonPayload,
	)

	if err != nil {
		return fmt.Errorf("failed to store event: %w", err)
	}

	return nil
}

func (r EventRepository) GetAllEvents(ctx context.Context) ([]StoredEvent, error) {
	var events []StoredEvent
	query := `
		SELECT event_id, published_at, event_name, event_payload
		FROM events
		ORDER BY published_at ASC
	`
	err := r.db.SelectContext(ctx, &events, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}
	return events, nil
}

func (r EventRepository) CountEvents(ctx context.Context) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM events")
	return count, err
}
