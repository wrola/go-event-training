package middleware

import (
	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
	"github.com/ThreeDotsLabs/watermill/message"
)

func HandleErrorMiddleware(next message.HandlerFunc) message.HandlerFunc {
	return func(msg *message.Message) ([]*message.Message, error) {
		res, err := next(msg)
		if err != nil {
			logger := log.FromContext(msg.Context())

			logger.With(
				"message_id", msg.UUID,
				"error", err.Error(),
			).Error("Error while handling a message")
		}
		return res, err
	}
}