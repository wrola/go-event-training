package commands

import (

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/redis/go-redis/v9"

	"tickets/commands/handler"
)




func NewCommandProcessor(
	redisClient *redis.Client,
	logger watermill.LoggerAdapter,
	receiptsService handler.ReceiptsService,
	paymentsService handler.PaymentsService,
	eventBus *cqrs.EventBus,
) (*message.Router, error) {

	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		return nil, err
	}

	cmdHandler := handler.NewCommandHandler(receiptsService, paymentsService, eventBus)

	commandProcessor, err := cqrs.NewCommandProcessorWithConfig(
		router,
		cqrs.CommandProcessorConfig{
			GenerateSubscribeTopic: func(params cqrs.CommandProcessorGenerateSubscribeTopicParams) (string, error) {
				return "commands." + params.CommandName, nil
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

		return redisstream.NewSubscriber(
			redisstream.SubscriberConfig{
				Client:        redisClient,
				ConsumerGroup: "commands." + params.HandlerName,
			}, logger)
	}
}
