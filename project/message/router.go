package message

import (
	"context"
	"encoding/json"
	"os"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/redis/go-redis/v9"
	"tickets/entities"
	"tickets/message/event"
)

type Router struct {
	router       *message.Router
	eventHandler *event.MessageHandler
	subReceipt   message.Subscriber
	subTracker   message.Subscriber
}

func NewRouter(eventHandler *event.MessageHandler) *Router {
	return &Router{
		eventHandler: eventHandler,
	}
}

func (r *Router) SetupSubscribers() error {
	logger := watermill.NewStdLogger(false, false)

	redisClient := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
	})

	var err error
	r.subReceipt, err = redisstream.NewSubscriber(
		redisstream.SubscriberConfig{
			Client:        redisClient,
			ConsumerGroup: "receipt-workers",
		}, logger)
	if err != nil {
		return err
	}

	r.subTracker, err = redisstream.NewSubscriber(
		redisstream.SubscriberConfig{
			Client:        redisClient,
			ConsumerGroup: "tracker-workers",
		}, logger)
	if err != nil {
		return err
	}

	return nil
}

func (r *Router) AddHandlers() error {
	logger := watermill.NewSlogLogger(nil)
	r.router = message.NewDefaultRouter(logger)

	r.router.AddConsumerHandler(
		"handler_ticket_receipt",
		"TicketBookingConfirmed",
		r.subReceipt,
		func(msg *message.Message) error {
			var event entities.TicketBookingConfirmed
			err := json.Unmarshal(msg.Payload, &event)
			if err != nil {
				return err
			}
			return r.eventHandler.IssueReceipt(msg.Context(), event)
		},
	)

	r.router.AddConsumerHandler(
		"handler_append_to_tracker",
		"TicketBookingConfirmed",
		r.subTracker,
		func(msg *message.Message) error {
			var event entities.TicketBookingConfirmed
			err := json.Unmarshal(msg.Payload, &event)
			if err != nil {
				return err
			}
			return r.eventHandler.AppendToTracker(msg.Context(), event)
		},
	)

	r.router.AddConsumerHandler(
		"handler_append_cancellation_to_tracker",
		"TicketBookingCanceled",
		r.subTracker,
		func(msg *message.Message) error {
			var event entities.TicketBookingCanceled
			err := json.Unmarshal(msg.Payload, &event)
			if err != nil {
				return err
			}
			return r.eventHandler.CancelTicket(msg.Context(), event)
		},
	)

	return nil
}

func (r *Router) Run(ctx context.Context) error {
	return r.router.Run(ctx)
}

func (r *Router) Running() <-chan struct{} {
	return r.router.Running()
}
