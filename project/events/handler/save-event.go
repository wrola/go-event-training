package handler

import (
	"encoding/json"
	"fmt"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
	"github.com/ThreeDotsLabs/watermill/message"
	"tickets/entities/events"
)

func (h *MessageHandler) SaveEventInEventStore(msg *message.Message) error {
	ctx := msg.Context()
	logger := log.FromContext(ctx)

	logger.Info("Saving event in event store")

	var envelope struct {
		Header events.MessageHeader `json:"header"`
	}

	if err := json.Unmarshal(msg.Payload, &envelope); err != nil {
		return fmt.Errorf("failed to unmarshal event header: %w", err)
	}

	
	eventName := msg.Metadata.Get("name")
	if eventName == "" {
		return fmt.Errorf("event name not found in message metadata")
	}

	return h.eventRepo.StoreEvent(
		ctx,
		envelope.Header.ID,
		envelope.Header.PublishedAt,
		eventName,
		msg.Payload,
	)
}
