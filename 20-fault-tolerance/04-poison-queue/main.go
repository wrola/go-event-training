package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
)

type orderStorage interface {
	AddTrackingLink(ctx context.Context, orderID string, trackingLink string) error
}

type OrderDispatched struct {
	OrderID      string `json:"order_id"`
	TrackingLink string `json:"tracking_link"`
}

type MissingTrackingLinkError struct {
	OrderID string
}

func (e MissingTrackingLinkError) Error() string {
	return fmt.Sprintf("order %s is missing tracking link", e.OrderID)
}

func ProcessMessages(
	ctx context.Context,
	sub message.Subscriber,
	pub message.Publisher,
	storage orderStorage,
) error {
	logger := watermill.NewSlogLogger(nil)
	router := message.NewDefaultRouter(logger)

	poisonQueueMiddleware, err := middleware.PoisonQueueWithFilter(
		pub,
		"PoisonQueue",
		func(err error) bool {
			var missingLinkErr MissingTrackingLinkError
			return errors.As(err, &missingLinkErr)
		},
	)
	if err != nil {
		return err
	}

	router.AddMiddleware(poisonQueueMiddleware)
	ep, err := cqrs.NewEventProcessorWithConfig(
		router,
		cqrs.EventProcessorConfig{
			GenerateSubscribeTopic: func(params cqrs.EventProcessorGenerateSubscribeTopicParams) (string, error) {
				return params.EventName, nil
			},
			SubscriberConstructor: func(params cqrs.EventProcessorSubscriberConstructorParams) (message.Subscriber, error) {
				return sub, nil
			},
			Marshaler: cqrs.JSONMarshaler{},
			Logger:    logger,
		},
	)
	if err != nil {
		return err
	}

	err = ep.AddHandlers(
		cqrs.NewEventHandler("OnOrderDispatched", func(ctx context.Context, event *OrderDispatched) error {
			// Validate that TrackingLink is present
			if event.TrackingLink == "" {
				return MissingTrackingLinkError{OrderID: event.OrderID}
			}

			// Store the tracking link (may fail with temporary errors like "database is down")
			return storage.AddTrackingLink(ctx, event.OrderID, event.TrackingLink)
		}),
	)
	if err != nil {
		return err
	}

	go func() {
		err := router.Run(ctx)
		if err != nil {
			panic(err)
		}
	}()

	<-router.Running()

	return nil
}
