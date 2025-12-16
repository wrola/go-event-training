package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	stdHTTP "net/http"

	"golang.org/x/sync/errgroup"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/clients"
	"github.com/ThreeDotsLabs/watermill/components/forwarder"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"

	"tickets/adapters"
	ticketsCommands "tickets/commands"
	"tickets/database"
	ticketsEvents "tickets/events"
	"tickets/events/handler"
	ticketsHttp "tickets/http"
	"tickets/metric"
)


type Service struct {
	echoRouter       *echo.Echo
	eventProcessor   *message.Router
	commandProcessor *message.Router
	forwarder        *forwarder.Forwarder
	db               *sqlx.DB
	migrator         *database.ReadModelMigrator
	traceProvider    *tracesdk.TracerProvider
}

func New(
	apiClients *clients.Clients,
	spreadsheetsAPI handler.SpreadsheetsAPI,
	receiptsService handler.ReceiptsService,
	paymentsService handler.PaymentsService,
	filesAPI handler.FilesAPI,
	deadNationAPI handler.DeadNationAPI,
	db *sqlx.DB,
) (*Service, error) {

	//TODO: refactor into container DI
	redisClient := ticketsEvents.NewRedisClient()
	logger := ticketsEvents.NewLogger()
	ticketRepository := database.NewTicketRepository(db)
	showsRepository := database.NewShowsRepository(db)
	bookingsRepository := database.NewBookingsRepository(db, logger)
	eventRepository := database.NewEventRepository(db)
	vipBundleRepository := database.NewVipBundleRepository(db, logger)

	var transportationService adapters.TransportationService
	if apiClients != nil {
		transportationService = adapters.NewTransportationServiceClient(apiClients)
	}

	publisher, err := ticketsEvents.NewMessagePublisher(redisClient, logger)
	if err != nil {
		return nil, err
	}

	eventBus := handler.NewEventBus(publisher)
	commandBus := handler.NewCommandBus(publisher)

	opsBookingRepository := database.NewOpsBookingReadModelRespository(db, eventBus)

	migrator := database.NewReadModelMigrator(eventRepository, opsBookingRepository)

	fwd, err := database.NewForwarder(db, publisher, logger)
	if err != nil {
		return nil, err
	}

	echoRouter, err := ticketsHttp.NewHttpRouter(
		eventBus,
		commandBus,
		ticketRepository,
		showsRepository,
		bookingsRepository,
		opsBookingRepository,
		vipBundleRepository,
	)
	if err != nil {
		return nil, err
	}

	outboxCollector := metric.NewOutboxCollector(db, database.OutboxTopic)
	prometheus.MustRegister(outboxCollector)

	eventProcessor, err := ticketsEvents.NewEventProcessor(
		receiptsService,
		spreadsheetsAPI,
		filesAPI,
		showsRepository,
		deadNationAPI,
		redisClient,
		logger,
		ticketRepository,
		eventBus,
		commandBus,
		opsBookingRepository,
		eventRepository,
		vipBundleRepository,
	)
	if err != nil {
		return nil, err
	}

	commandProcessor, err := ticketsCommands.NewCommandProcessor(
		redisClient,
		logger,
		receiptsService,
		paymentsService,
		transportationService,
		eventBus,
		bookingsRepository,
	)

	if err != nil {
		return nil, err
	}

	return &Service{
		echoRouter:       echoRouter,
		eventProcessor:   eventProcessor,
		commandProcessor: commandProcessor,
		forwarder:        fwd,
		db:               db,
		migrator:         migrator,
		traceProvider:    nil,
	}, nil
}

func (s *Service) SetTraceProvider(tp *tracesdk.TracerProvider) {
	s.traceProvider = tp
}

func (s *Service) Run(ctx context.Context) error {
	database.InitializeSchema(s.db)

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return s.eventProcessor.Run(ctx)
	})

	g.Go(func() error {
		// Wait for event processor to be ready
		<-s.eventProcessor.Running()

		if err := s.migrator.MigrateReadModel(ctx); err != nil {
			fmt.Printf("ERROR: Read model migration failed: %v\n", err)
			return err
		}

		fmt.Println("Read model migration completed successfully")
		return nil
	})

	g.Go(func() error {
		return s.commandProcessor.Run(ctx)
	})

	g.Go(func() error {
		return s.forwarder.Run(ctx)
	})

	g.Go(func() error {
		<-s.eventProcessor.Running()
		<-s.commandProcessor.Running()
		<-s.forwarder.Running()

		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}

		err := s.echoRouter.Start(":" + port)
		if err != nil && !errors.Is(err, stdHTTP.ErrServerClosed) {
			return err
		}

		return nil
	})

	g.Go(func() error {
		<-ctx.Done()
		return s.echoRouter.Shutdown(context.Background())
	})

	g.Go(func() error {
		<-ctx.Done()
		if s.traceProvider != nil {
			return s.traceProvider.Shutdown(context.Background())
		}
		return nil
	})

	return g.Wait()
}

func (s *Service) RunWithGracefulShutdown(port string) error {
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer cancel()

	if port != "" {
		os.Setenv("PORT", port)
	}

	return s.Run(ctx)
}
