package events

import (
	"fmt"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/redis/go-redis/v9"
	"tickets/database"
	"tickets/events/handler"
	ticketsMiddleware "tickets/events/middleware"
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
	eventRepository database.EventRepository,
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
		eventRepository,
	)

	redisSubscriber := NewSubscriberConstructor(redisClient, logger)
	redisPublisher, err := NewMessagePublisher(redisClient, logger)
	if err != nil {
		return nil, err
	}

	eventProcessor, err := cqrs.NewEventProcessorWithConfig(
		router,
		cqrs.EventProcessorConfig{
			GenerateSubscribeTopic: func(params cqrs.EventProcessorGenerateSubscribeTopicParams) (string, error) {
				if len(params.EventName) >= 8 && params.EventName[:8] == "Internal" {
					return "internal-events.svc-tickets." + params.EventName, nil
				}

				return "events." + params.EventName, nil
			},
			SubscriberConstructor: redisSubscriber,
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
			"OnTicketBookingConfirmed_v1",
			msgHandler.OnTicketBookingConfirmed_v1,
		),
		cqrs.NewEventHandler(
			"OnTicketReceiptIssued_v1",
			msgHandler.OnTicketReceiptIssued_v1,
		),
		cqrs.NewEventHandler(
			"OnTicketPrinted_v1",
			msgHandler.OnTicketPrinted_v1,
		),
		cqrs.NewEventHandler(
			"OnTicketRefunded_v1",
			msgHandler.OnTicketRefunded_v1,
		),
		cqrs.NewEventHandler(
			"UpdateRefundedTicket",
			msgHandler.UpdateRefundedTicket,
		),
	)
	if err != nil {
		return nil, err
	}

	eventProcessorMarshaler := cqrs.JSONMarshaler{
		GenerateName: cqrs.StructName,
	}

	splitterSubscriber, err := redisstream.NewSubscriber(
		redisstream.SubscriberConfig{
			Client:        redisClient,
			ConsumerGroup: "events_splitter",
		}, logger)
	if err != nil {
		return nil, err
	}

	router.AddConsumerHandler(
		"events_splitter",
		"events",
		splitterSubscriber,
		func(msg *message.Message) error {
			eventName := eventProcessorMarshaler.NameFromMessage(msg)
			if eventName == "" {
				return fmt.Errorf("cannot get event name from message")
			}

			return redisPublisher.Publish("events."+eventName, msg)
		},
	)

	eventStoreSubscriber, err := redisstream.NewSubscriber(
		redisstream.SubscriberConfig{
			Client:        redisClient,
			ConsumerGroup: "event_store",
		}, logger)
	if err != nil {
		return nil, err
	}

	router.AddConsumerHandler(
		"event_store_handler",
		"events",
		eventStoreSubscriber,
		msgHandler.SaveEventInEventStore,
	)

	return router, nil
}

func SetupMiddlewares(router *message.Router) {
	router.AddMiddleware(ticketsMiddleware.AttachCorrelationIdMiddleware)
	router.AddMiddleware(ticketsMiddleware.LogMessagesMiddleware)
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