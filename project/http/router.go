package http

import (
	libHttp "github.com/ThreeDotsLabs/go-event-driven/v2/common/http"
	"github.com/labstack/echo/v4"

	"tickets/message"
)

func NewHttpRouter() (*echo.Echo, error) {
	e := libHttp.NewEcho()

	publisher, err := message.NewMessageProducer()
	if err != nil {
		return nil, err
	}

	handler := Handler{
		publisher: publisher,
	}

	e.POST("/tickets-confirmation", handler.PostTicketsConfirmation)

	return e, nil
}
