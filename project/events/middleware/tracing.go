package middleware

import (
	"fmt"

	"github.com/ThreeDotsLabs/watermill/message"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
)

func TracingMiddleware(next message.HandlerFunc) message.HandlerFunc {
	return func(msg *message.Message) (events []*message.Message, err error) {
		ctx := msg.Context()

		ctx = otel.GetTextMapPropagator().Extract(
			ctx,
			propagation.MapCarrier(msg.Metadata),
		)

		topic := message.SubscribeTopicFromCtx(ctx)
		handlerName := message.HandlerNameFromCtx(ctx)

		ctx, span := otel.Tracer("").Start(
			ctx,
			fmt.Sprintf("Handler: %s", handlerName),
		)

		span.SetAttributes(
			attribute.String("topic", topic),
			attribute.String("handler", handlerName),
			attribute.String("message_id", msg.UUID),
		)

		defer func() {
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
			}
			span.End()
		}()

		msg.SetContext(ctx)

		events, err = next(msg)

		return events, err
	}
}
