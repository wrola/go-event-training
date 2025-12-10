package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

type PaymentReceived struct {
	ID            string `json:"id"`
	RoomBookingID string `json:"room_booking_id"`
}

type RoomBookingConfirmed struct {
	RoomBookingID string `json:"room_booking_id"`
}

func NewRouter(rdb *redis.Client, logger watermill.LoggerAdapter) (*message.Router, *cqrs.EventBus) {
	router := message.NewDefaultRouter(logger)

	router.AddMiddleware(func(h message.HandlerFunc) message.HandlerFunc {
		return func(msg *message.Message) (events []*message.Message, err error) {
			ctx := msg.Context()

			ctx = otel.GetTextMapPropagator().Extract(ctx, propagation.MapCarrier(msg.Metadata))

			topic := message.SubscribeTopicFromCtx(ctx)
			handlerName := message.HandlerNameFromCtx(ctx)

			ctx, span := otel.Tracer("").Start(ctx, fmt.Sprintf("Handler: %s", handlerName))

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

			events, err = h(msg)

			return events, err
		}
	})

	var pub message.Publisher
	pub, err := redisstream.NewPublisher(
		redisstream.PublisherConfig{
			Client: rdb,
		},
		logger,
	)
	if err != nil {
		panic(err)
	}

	pub = newTracePublisher(pub)
	
	marshaler := cqrs.JSONMarshaler{
		GenerateName: cqrs.StructName,
	}
	eventBus, err := cqrs.NewEventBusWithConfig(
		pub,
		cqrs.EventBusConfig{
			GeneratePublishTopic: func(params cqrs.GenerateEventPublishTopicParams) (string, error) {
				return params.EventName, nil
			},
			Marshaler: marshaler,
		},
	)
	if err != nil {
		panic(err)
	}

	processor, err := newEventProcessor(router, rdb, marshaler, logger)
	if err != nil {
		panic(err)
	}

	err = processor.AddHandlers(cqrs.NewEventHandler(
		"PaymentReceived",
		func(ctx context.Context, event *PaymentReceived) error {
			return eventBus.Publish(ctx, RoomBookingConfirmed{
				RoomBookingID: event.RoomBookingID,
			})
		},
	))
	if err != nil {
		panic(err)
	}

	return router, eventBus
}

func initTracing(exp sdktrace.SpanExporter) {
	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("ExampleService"),
		),
	)
	// we are ignoring schema conflicts, more here: https://github.com/open-telemetry/opentelemetry-go/pull/4876
	if err != nil && !errors.Is(err, resource.ErrSchemaURLConflict) {
		panic(err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exp),
		sdktrace.WithResource(r),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})
}

func newEventProcessor(
	router *message.Router,
	rdb *redis.Client,
	marshaler cqrs.CommandEventMarshaler,
	logger watermill.LoggerAdapter,
) (*cqrs.EventProcessor, error) {
	return cqrs.NewEventProcessorWithConfig(
		router,
		cqrs.EventProcessorConfig{
			SubscriberConstructor: func(params cqrs.EventProcessorSubscriberConstructorParams) (message.Subscriber, error) {
				return redisstream.NewSubscriber(
					redisstream.SubscriberConfig{
						Client:        rdb,
						ConsumerGroup: "svc-something." + params.HandlerName,
					},
					logger,
				)
			},
			GenerateSubscribeTopic: func(params cqrs.EventProcessorGenerateSubscribeTopicParams) (string, error) {
				return params.EventName, nil
			},
			Marshaler: marshaler,
			Logger:    logger,
		},
	)
}

type TracedPublisher struct {
	wrapped message.Publisher
}

func (p *TracedPublisher) Publish(topic string, messages ...*message.Message) error {
	for i := range messages {
	otel.GetTextMapPropagator().Inject(messages[i].Context(), propagation.MapCarrier(messages[i].Metadata))
	}
	return p.wrapped.Publish(topic, messages...)
}

func (p *TracedPublisher) Close() error {
	return p.wrapped.Close()
}

func newTracePublisher(pub message.Publisher) message.Publisher {
	return &TracedPublisher{wrapped: pub}
}