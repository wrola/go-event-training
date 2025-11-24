package service

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"golang.org/x/sync/errgroup"
	stdHTTP "net/http"

	"github.com/labstack/echo/v4"

	ticketsHttp "tickets/http"
	"tickets/message"
	"tickets/message/event"
)

type Service struct {
	echoRouter     *echo.Echo
	messageRouter  *message.Router
}

func New(
	spreadsheetsAPI event.SpreadsheetsAPI,
	receiptsService event.ReceiptsService,
) (Service, error) {
	echoRouter, err := ticketsHttp.NewHttpRouter()
	if err != nil {
		return Service{}, err
	}

	// Create event handler
	eventHandler := event.NewMessageHandler(spreadsheetsAPI, receiptsService)

	// Create router with the event handler
	messageRouter := message.NewRouter(eventHandler)

	// Setup subscribers
	err = messageRouter.SetupSubscribers()
	if err != nil {
		return Service{}, err
	}

	service := Service{
		echoRouter:    echoRouter,
		messageRouter: messageRouter,
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	g, ctx := errgroup.WithContext(ctx)

	// Add handlers to the router
	err = messageRouter.AddHandlers()
	if err != nil {
		cancel()
		return Service{}, err
	}

	// Start processing messages in a separate goroutine
	g.Go(func() error {
		return messageRouter.Run(ctx)
	})

	g.Go(func() error {
		<-messageRouter.Running()

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
	<-s.messageRouter.Running()
	port := os.Getenv("PORT")
	err := s.echoRouter.Start(":" + port)

	if err != nil && !errors.Is(err, stdHTTP.ErrServerClosed) {
		return err
	}
	return nil
}
