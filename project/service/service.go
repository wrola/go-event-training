package service

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"golang.org/x/sync/errgroup"
	stdHTTP "net/http"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/labstack/echo/v4"

	ticketsHttp "tickets/http"
	ticketsMessage "tickets/message"
	"tickets/message/event"
)

type Service struct {
	echoRouter      *echo.Echo
	eventProcessor  *message.Router
}

func NewService(
	spreadsheetsAPI event.SpreadsheetsAPI,
	receiptsService event.ReceiptsService,
) (Service, error) {

	redisClient := ticketsMessage.NewRedisClient()
	logger := ticketsMessage.NewLogger()

	publisher, err := ticketsMessage.NewMessagePublisher(redisClient, logger)
	if err != nil {
		return Service{}, err
	}

	eventBus := event.NewEventBus(publisher)

	echoRouter, err := ticketsHttp.NewHttpRouter(eventBus)
	if err != nil {
		return Service{}, err
	}

	eventProcessor, err := ticketsMessage.NewEventProcessor(receiptsService, spreadsheetsAPI, redisClient, logger)
	if err != nil {
		return Service{}, err
	}

	service := Service{
		echoRouter:     echoRouter,
		eventProcessor: eventProcessor,
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return eventProcessor.Run(ctx)
	})

	g.Go(func() error {
		<-eventProcessor.Running()

		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}

		err := echoRouter.Start(":" + port)
		if err != nil && !errors.Is(err, stdHTTP.ErrServerClosed) {
			return err
		}

		return nil
	})

	g.Go(func() error {
		<-ctx.Done()
		cancel()
		return echoRouter.Shutdown(ctx)
	})

	err = g.Wait()
	if err != nil {
		return Service{}, err
	}

	return service, nil
}

func (s Service) Run(ctx context.Context) error {
	<-s.eventProcessor.Running()
	port := os.Getenv("PORT")
	err := s.echoRouter.Start(":" + port)

	if err != nil && !errors.Is(err, stdHTTP.ErrServerClosed) {
		return err
	}
	return nil
}
