package service

import (
	"context"
	"errors"
	stdHTTP "net/http"

	"github.com/labstack/echo/v4"

	ticketsHttp "tickets/http"
	"tickets/worker"
)

type Service struct {
	echoRouter *echo.Echo
	worker    *worker.Worker
}

func New(
	spreadsheetsAPI worker.SpreadsheetsAPI,
	receiptsService worker.ReceiptsService,
) Service {
	w := worker.NewWorker(spreadsheetsAPI, receiptsService)
	echoRouter := ticketsHttp.NewHttpRouter(w)
		

	return Service{
		echoRouter: echoRouter,
		worker:    w,
	}
}

func (s Service) Run(ctx context.Context) error {
	go func() {
		s.worker.Run(ctx)
	}()

	err := s.echoRouter.Start(":8080")
	if err != nil && !errors.Is(err, stdHTTP.ErrServerClosed) {
		return err
	}

	return nil
}
