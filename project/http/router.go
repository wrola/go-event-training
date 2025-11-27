package http

import (
	"net/http"

	libHttp "github.com/ThreeDotsLabs/go-event-driven/v2/common/http"
	"github.com/labstack/echo/v4"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
)

func NewHttpRouter(eventBus *cqrs.EventBus) (*echo.Echo, error) {
	e := libHttp.NewEcho()

	handler := Handler{
		eventBus: eventBus,
	}

	e.POST("/tickets-status", handler.PostTicketsConfirmation)
	e.GET("/health", func(c echo.Context) error {
	
		return c.String(http.StatusOK, "ok")
	})

	return e, nil
}
