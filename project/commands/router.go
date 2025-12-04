package commands

import (
	"context"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/redis/go-redis/v9"

	"tickets/commands/handler"
	"tickets/entities/events"
)

type ReceiptsService interface {
	IssueReceipt(ctx context.Context, payload events.TicketBookingConfirmed) error
	VoidReceipt(ctx context.Context, ticketID string, reason string, idempotencyKey string) error
}

func NewCommandProcessor(
	redisClient *redis.Client,
	logger watermill.LoggerAdapter,
	receiptsService ReceiptsService,
) (*message.Router, error) {

	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		return nil, err
	}

	cmdHandler := handler.NewCommandHandler(receiptsService)

	commandProcessor, err := cqrs.NewCommandProcessorWithConfig(
		router,
		cqrs.CommandProcessorConfig{
			GenerateSubscribeTopic: func(params cqrs.CommandProcessorGenerateSubscribeTopicParams) (string, error) {
				return params.CommandName, nil
			},
			SubscriberConstructor: NewSubscriberConstructor(redisClient, logger),
			Marshaler: cqrs.JSONMarshaler{
				GenerateName: cqrs.StructName,
			},
			Logger: logger,
		},
	)
	if err != nil {
		return nil, err
	}

	err = commandProcessor.AddHandlers(
		cqrs.NewCommandHandler(
			"RefundTicketHandler",
			cmdHandler.RefundTicket,
		),
	)
	if err != nil {
		return nil, err
	}

	return router, nil
}

func NewSubscriberConstructor(redisClient *redis.Client, logger watermill.LoggerAdapter) cqrs.CommandProcessorSubscriberConstructorFn {
	return func(params cqrs.CommandProcessorSubscriberConstructorParams) (message.Subscriber, error) {
		consumerGroup := params.HandlerName + "-workers"

		return redisstream.NewSubscriber(
			redisstream.SubscriberConfig{
				Client:        redisClient,
				ConsumerGroup: consumerGroup,
			}, logger)
	}
}
