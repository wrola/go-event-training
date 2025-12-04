package events

import (
	"time"

	"github.com/google/uuid"
)

type MessageHeader struct {
	ID             string    `json:"id"`
	PublishedAt    time.Time `json:"published_at"`
	IdempotencyKey string    `json:"idempotency_key"`
}

func NewMessageHeader() MessageHeader {
	return MessageHeader{
		ID:             uuid.NewString(),
		PublishedAt:    time.Now().UTC(),
		IdempotencyKey: uuid.NewString(),
	}
}

func NewMessageHeaderWithIdempotencyKey(idempotencyKey string) MessageHeader {
	return MessageHeader{
		ID:             uuid.NewString(),
		PublishedAt:    time.Now().UTC(),
		IdempotencyKey: idempotencyKey,
	}
}
