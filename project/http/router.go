package http

import (
	libHttp "github.com/ThreeDotsLabs/go-event-driven/v2/common/http"
	"github.com/labstack/echo/v4"
	"tickets/worker"
)

func NewHttpRouter(w *worker.Worker ) *echo.Echo {
	e := libHttp.NewEcho()

	handler := Handler{
		Worker: w,
	}

	e.POST("/tickets-confirmation", handler.PostTicketsConfirmation)

	return e
}
