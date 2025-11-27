package http

import (
	"net/http"
	"github.com/labstack/echo/v4"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"tickets/entities"
)

type Handler struct {
	eventBus *cqrs.EventBus
}

type TicketStatusRequest struct {
	TicketID string `json:"ticket_id"`
	CustomerEmail string `json:"customer_email"`
	Price entities.Money `json:"price"`
	Status entities.TicketBookingStatus `json:"status"`
}

type ticketsConfirmationRequest struct {
	Tickets []TicketStatusRequest `json:"tickets"`
}

func (h Handler) PostTicketsConfirmation(c echo.Context) error {
	var request ticketsConfirmationRequest
	if err := c.Bind(&request);  err != nil {
		return err
	}

	for _, ticket := range request.Tickets {
  	switch ticket.Status {
  		case entities.TicketStatusConfirmed:
  			event := entities.NewTicketBookingConfirmed(ticket.TicketID, ticket.CustomerEmail, ticket.Price)
  			h.eventBus.Publish(c.Request().Context(), event)
  		case entities.TicketStatusCanceled:
  			event := entities.NewTicketBookingCanceled(ticket.TicketID, ticket.CustomerEmail, ticket.Price)
  			h.eventBus.Publish(c.Request().Context(), event)
  		default:
  			return echo.NewHTTPError(http.StatusBadRequest, "invalid ticket status")
  	}
  }

	return c.NoContent(http.StatusOK)
}
