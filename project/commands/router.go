package commands

import (
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/redis/go-redis/v9"

	"tickets/adapters"
	"tickets/commands/handler"
	"tickets/database"
	ticketsMiddleware "tickets/events/middleware"
)




func NewCommandProcessor(
	redisClient *redis.Client,
	logger watermill.LoggerAdapter,
	receiptsService handler.ReceiptsService,
	paymentsService handler.PaymentsService,
	transportationService adapters.TransportationService,
	eventBus *cqrs.EventBus,
	bookingsRepository database.BookingsRepository,
) (*message.Router, error) {

	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		return nil, err
	}

	SetupMiddlewares(router)

	cmdHandler := handler.NewCommandHandler(receiptsService, paymentsService, transportationService, eventBus, bookingsRepository)

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
		cqrs.NewCommandHandler(
			"BookShowTicketsHandler",
			cmdHandler.BookShowTickets,
		),
		cqrs.NewCommandHandler(
			"BookFlightHandler",
			cmdHandler.BookFlight,
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

func SetupMiddlewares(router *message.Router) {
	router.AddMiddleware(ticketsMiddleware.AttachCorrelationIdMiddleware)
	router.AddMiddleware(ticketsMiddleware.TracingMiddleware)
	router.AddMiddleware(ticketsMiddleware.MetricsMiddleware)
	router.AddMiddleware(ticketsMiddleware.SkipPermanentErrorsMiddleware)
	router.AddMiddleware(ticketsMiddleware.HandleErrorMiddleware)
	router.AddMiddleware(middleware.Retry{
		MaxRetries:      10,
		InitialInterval: time.Millisecond * 100,
		MaxInterval:     time.Second,
		Multiplier:      2,
	}.Middleware)
}
