package service

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"golang.org/x/sync/errgroup"
	stdHTTP "net/http"

	"github.com/labstack/echo/v4"

	ticketsHttp "tickets/http"
	"tickets/message"
)

type Service struct {
	echoRouter     *echo.Echo
	messageHandler *message.MessageHandler
}

func New(
	spreadsheetsAPI message.SpreadsheetsAPI,
	receiptsService message.ReceiptsService,
) (Service, error) {
	echoRouter, err := ticketsHttp.NewHttpRouter()
	if err != nil {
		return Service{}, err
	}

	messageHandler := message.NewMessageHandler(spreadsheetsAPI, receiptsService)

	err = messageHandler.NewMessageConsumers()
	if err != nil {
		return Service{}, err
	}

	service := Service{
		echoRouter:     echoRouter,
		messageHandler: messageHandler,
	}

	ctx, cancel := signal.NotifyContext(context.Background())

	g, ctx := errgroup.WithContext(ctx)

	// Initialize message consumers (creates the router)
	err = messageHandler.AddConsumers(ctx)
	if err != nil {
		cancel()
		return Service{}, err
	}

	// Start processing messages in a separate goroutine
	g.Go(func() error {
		return messageHandler.StartProcessMessages(ctx)
	})

	g.Go(func() error {
		<-messageHandler.Running()

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
	<-s.messageHandler.Running()
	port := os.Getenv("PORT")
	err := s.echoRouter.Start(":" + port)

	if err != nil && !errors.Is(err, stdHTTP.ErrServerClosed) {
		return err
	}
	return nil
}
