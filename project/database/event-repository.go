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
