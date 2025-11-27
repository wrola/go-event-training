package http

import (
	"net/http"

	libHttp "github.com/ThreeDotsLabs/go-event-driven/v2/common/http"
	"github.com/labstack/echo/v4"
	"tickets/message"
)

func NewHttpRouter() (*echo.Echo, error) {
	e := libHttp.NewEcho()

	pub, err := message.NewMessageProducer()
	if err != nil {
		return nil, err
	}

	eventBus, errEventBus := message.NewEventBus(pub)

	if errEventBus != nil {
		return nil, errEventBus
	}


	handler := Handler{
		eventBus: eventBus,
	}

	e.POST("/tickets-status", handler.PostTicketsConfirmation)
	e.GET("/health", func(c echo.Context) error {
	
		return c.String(http.StatusOK, "ok")
	})
	return e, nil
}
