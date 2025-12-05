package service

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"
	stdHTTP "net/http"

	"github.com/ThreeDotsLabs/watermill/components/forwarder"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/labstack/echo/v4"
	"github.com/jmoiron/sqlx"

	ticketsHttp "tickets/http"
	ticketsEvents "tickets/events"
	"tickets/events/handler"
	ticketsCommands "tickets/commands"
	"tickets/database"
)

type Service struct {
	echoRouter       *echo.Echo
	eventProcessor   *message.Router
	commandProcessor *message.Router
	forwarder        *forwarder.Forwarder
	db               *sqlx.DB
}

func New(
	spreadsheetsAPI handler.SpreadsheetsAPI,
	receiptsService handler.ReceiptsService,
	paymentsService handler.PaymentsService,
	filesAPI handler.FilesAPI,
	deadNationAPI handler.DeadNationAPI,
	db *sqlx.DB,
) (*Service, error) {

	redisClient := ticketsEvents.NewRedisClient()
	logger := ticketsEvents.NewLogger()
	ticketRepository := database.NewTicketRepository(db)
	showsRepository := database.NewShowsRepository(db)
	bookingsRepository := database.NewBookingsRepository(db, logger)

	publisher, err := ticketsEvents.NewMessagePublisher(redisClient, logger)
	if err != nil {
		return nil, err
	}

	fwd, err := database.NewForwarder(db, publisher, logger)
	if err != nil {
		return nil, err
	}

	eventBus := handler.NewEventBus(publisher)
	commandBus := handler.NewCommandBus(publisher)

	echoRouter, err := ticketsHttp.NewHttpRouter(eventBus, commandBus, ticketRepository, showsRepository, bookingsRepository)
	if err != nil {
		return nil, err
	}

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
	)
	if err != nil {
		return nil, err
	}

	commandProcessor, err := ticketsCommands.NewCommandProcessor(
		redisClient,
		logger,
		receiptsService,
		paymentsService,
		eventBus,
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
	}, nil
}

func (s *Service) Run(ctx context.Context) error {
	database.InitializeSchema(s.db)

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return s.eventProcessor.Run(ctx)
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
