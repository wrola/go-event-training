package message

import (
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/redis/go-redis/v9"
	"tickets/message/event"
	ticketsMiddleware "tickets/message/middleware"
)

func NewEventProcessor(receiptsService event.ReceiptsService, spreadsheetsAPI event.SpreadsheetsAPI, redisClient *redis.Client, logger watermill.LoggerAdapter) (*message.Router, error) {

	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		return nil, err
	}


	handler := event.NewMessageHandler(spreadsheetsAPI, receiptsService)

	// Create EventProcessor
	eventProcessor, err := cqrs.NewEventProcessorWithConfig(
		router,
		cqrs.EventProcessorConfig{
			GenerateSubscribeTopic: func(params cqrs.EventProcessorGenerateSubscribeTopicParams) (string, error) {
				return params.EventName, nil
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

	// Register event handlers
	err = eventProcessor.AddHandlers(
		cqrs.NewEventHandler(
			"IssueReceipt",
			handler.IssueReceipt,
		),
		cqrs.NewEventHandler(
			"AppendToTracker",
			handler.AppendToTracker,
		),
		cqrs.NewEventHandler(
			"CancelTicket",
			handler.CancelTicket,
		),
	)
	if err != nil {
		return nil, err
	}

	return router, nil
}

func SetupMiddlewares(router *message.Router) { 
	router.AddMiddleware(ticketsMiddleware.AttachCorrelationIdMiddleware)
	router.AddMiddleware(ticketsMiddleware.LogMessagesMiddleware)
	router.AddMiddleware(ticketsMiddleware.SkipPermanentErrorsMiddleware)
	router.AddMiddleware(ticketsMiddleware.HandleErrorMiddleware)
	router.AddMiddleware(middleware.Retry{
		MaxRetries:      10,
		InitialInterval: time.Millisecond * 100,
		MaxInterval:     time.Second,
		Multiplier:      2,
	}.Middleware)
}