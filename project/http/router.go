package http

import (
	"net/http"

	libHttp "github.com/ThreeDotsLabs/go-event-driven/v2/common/http"
	"github.com/labstack/echo/v4"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"tickets/database"
)

func NewHttpRouter(eventBus *cqrs.EventBus, commandBus *cqrs.CommandBus, ticketRepository database.TicketRepository, showsRepository database.ShowsRepository, bookingsRepository database.BookingsRepository, opsBookingRepository database.OpsBookingReadModelRepository) (*echo.Echo, error) {
	e := libHttp.NewEcho()

	handler := NewHandler(eventBus, commandBus, ticketRepository, showsRepository, bookingsRepository, opsBookingRepository)

	e.GET("/tickets", handler.GetAllTickets)
	e.POST("/tickets-status", handler.PostTicketsConfirmation)
	e.POST("/shows", handler.CreateShow)
	e.POST("/book-tickets", handler.BookTickets)
	e.PUT("/ticket-refund/:ticket_id", handler.PutTicketRefund)
	e.GET("/ops/bookings", handler.GetOpsBookings)
	e.GET("/ops/bookings/:booking_id", handler.GetOpsBookingByID)
	e.GET("/health", func(c echo.Context) error {

		return c.String(http.StatusOK, "ok")
	})

	return e, nil
}
