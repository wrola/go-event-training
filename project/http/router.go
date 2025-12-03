package http

import (
	"net/http"

	libHttp "github.com/ThreeDotsLabs/go-event-driven/v2/common/http"
	"github.com/labstack/echo/v4"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"tickets/database"
)

func NewHttpRouter(eventBus *cqrs.EventBus, ticketRepository database.TicketRepository, showsRepository database.ShowsRepository, bookingsRepository database.BookingsRepository) (*echo.Echo, error) {
	e := libHttp.NewEcho()

	handler := NewHandler(eventBus, ticketRepository, showsRepository, bookingsRepository)

	e.GET("/tickets", handler.GetAllTickets)
	e.POST("/tickets-status", handler.PostTicketsConfirmation)
	e.POST("/shows", handler.CreateShow)
	e.POST("/book-tickets", handler.BookTickets)
	e.GET("/health", func(c echo.Context) error {
	
		return c.String(http.StatusOK, "ok")
	})

	return e, nil
}
