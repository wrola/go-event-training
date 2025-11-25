package middleware

import (
	"log/slog"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
)

func AttachCorrelationIdMiddleware(next message.HandlerFunc) message.HandlerFunc {
	return func(msg *message.Message) ([]*message.Message, error) {
		correlationID := middleware.MessageCorrelationID(msg)

		ctx := msg.Context()

		ctx = log.ContextWithCorrelationID(ctx, correlationID)

		logger := slog.With("correlation_id", correlationID)
		ctx = log.ToContext(ctx, logger)

		msg.SetContext(ctx)

		return next(msg)
	}
}
