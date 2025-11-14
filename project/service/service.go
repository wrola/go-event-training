package service

import (
	"context"
	"errors"
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

	subRecepit, subTracker, err := message.NewMessageConsumers()
	if err != nil {
		return Service{}, err
	}

	service := Service{
		echoRouter:     echoRouter,
		messageHandler: messageHandler,
	}

	go messageHandler.RunSubscribers(context.Background(), subRecepit, subTracker)

	return service, nil
}

func (s Service) Run(ctx context.Context) error {
	err := s.echoRouter.Start(":8080")
	if err != nil && !errors.Is(err, stdHTTP.ErrServerClosed) {
		return err
	}

	return nil
}
