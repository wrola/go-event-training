package events

import (
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/redis/go-redis/v9"
	"tickets/events/handler"
	ticketsMiddleware "tickets/events/middleware"
	"tickets/database"
)

func NewEventProcessor(
	receiptsService handler.ReceiptsService,
	spreadsheetsAPI handler.SpreadsheetsAPI,
	filesAPI handler.FilesAPI,
	showsRepository database.ShowsRepository,
	deadNationAPI handler.DeadNationAPI,
	redisClient *redis.Client,
	logger watermill.LoggerAdapter,
	ticketRepository database.TicketRepository,
	eventBus *cqrs.EventBus,
	opsBookingRepository database.OpsBookingReadModelRepository,
) (*message.Router, error) {

	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		return nil, err
	}

	SetupMiddlewares(router)

	msgHandler := handler.NewMessageHandler(
		spreadsheetsAPI,
		receiptsService,
		ticketRepository,
		filesAPI,
		showsRepository,
		deadNationAPI,
		eventBus,
		opsBookingRepository,
	)

	eventProcessor, err := cqrs.NewEventProcessorWithConfig(
		router,
		cqrs.EventProcessorConfig{
			GenerateSubscribeTopic: func(params cqrs.EventProcessorGenerateSubscribeTopicParams) (string, error) {
				return "events." + params.EventName, nil
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

	err = eventProcessor.AddHandlers(
		cqrs.NewEventHandler(
			"IssueReceipt",
			msgHandler.IssueReceipt,
		),
		cqrs.NewEventHandler(
			"AppendToTracker",
			msgHandler.AppendToTracker,
		),
		cqrs.NewEventHandler(
			"CancelTicket",
			msgHandler.CancelTicket,
		),
		cqrs.NewEventHandler(
			"StoreTicket",
			msgHandler.StoreTicket,
		),
		cqrs.NewEventHandler(
			"RemoveCancledTicket",
			msgHandler.RemoveCanceledTicket,
		),
		cqrs.NewEventHandler(
			"StoreTicketFile",
			msgHandler.StoreTicketFile,
		),
		cqrs.NewEventHandler(
			"BookingMade",
			msgHandler.BookingMadeUpdate,
		),
		cqrs.NewEventHandler(
			"OnBookingMade",
			msgHandler.OnBookingMade,
		),
		cqrs.NewEventHandler(
			"OnTicketBookingConfirmed",
			msgHandler.OnTicketBookingConfirmed,
		),
		cqrs.NewEventHandler(
			"OnTicketReceiptIssued",
			msgHandler.OnTicketReceiptIssued,
		),
		cqrs.NewEventHandler(
			"OnTicketPrinted",
			msgHandler.OnTicketPrinted,
		),
		cqrs.NewEventHandler(
			"OnTicketRefunded",
			msgHandler.OnTicketRefunded,
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