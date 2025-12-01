package http

import (
	"net/http"

	libHttp "github.com/ThreeDotsLabs/go-event-driven/v2/common/http"
	"github.com/labstack/echo/v4"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"tickets/database"
)

func NewHttpRouter(eventBus *cqrs.EventBus, ticketRepository database.TicketRepository) (*echo.Echo, error) {
	e := libHttp.NewEcho()

	handler := NewHandler(eventBus, ticketRepository)

	e.GET("/tickets", handler.GetAllTickets)
	e.POST("/tickets-status", handler.PostTicketsConfirmation)
	e.GET("/health", func(c echo.Context) error {
	
		return c.String(http.StatusOK, "ok")
	})

	return e, nil
}
