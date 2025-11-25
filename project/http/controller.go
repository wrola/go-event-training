package http

import (
	"encoding/json"
	"net/http"
	"github.com/labstack/echo/v4"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"tickets/entities"
)

type Handler struct {
	publisher message.Publisher
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
  			eventPayload, err := json.Marshal(event)

  			if err != nil {
  				return err
  			}

  			msg := message.NewMessage(watermill.NewUUID(), eventPayload)
			msg.Metadata.Set("correlation_id", c.Request().Header.Get("Correlation-ID"))
  			h.publisher.Publish("TicketBookingConfirmed", msg)
  		case entities.TicketStatusCanceled:
  			event := entities.NewTicketBookingCanceled(ticket.TicketID, ticket.CustomerEmail, ticket.Price)
  			eventPayload, err := json.Marshal(event)

  			if err != nil {
  				return err
  			}

  			msg := message.NewMessage(watermill.NewUUID(), eventPayload)
			msg.Metadata.Set("correlation_id", c.Request().Header.Get("Correlation-ID"))
  			h.publisher.Publish("TicketBookingCanceled", msg)
  		default:
  			return echo.NewHTTPError(http.StatusBadRequest, "invalid ticket status")
  	}
  }

	return c.NoContent(http.StatusOK)
}
