package entities

import (
	"time"

	"github.com/google/uuid"
)

type MessageHeader struct {
	ID          string    `json:"id"`
	PublishedAt time.Time `json:"published_at"`
}

func NewMessageHeader() MessageHeader {
	return MessageHeader{
		ID:          uuid.NewString(),
		PublishedAt: time.Now().UTC(),
	}
}
