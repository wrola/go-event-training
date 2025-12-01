package http

import (
	"net/http"
	"github.com/labstack/echo/v4"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"tickets/database"
	"tickets/entities"
)

type Handler struct {
	eventBus         *cqrs.EventBus
	ticketRepository database.TicketRepository
}

func NewHandler(eventBus *cqrs.EventBus, ticketRepository database.TicketRepository) Handler {
	if eventBus == nil {
		panic("eventBus is required")
	}

	return Handler{
		eventBus:         eventBus,
		ticketRepository: ticketRepository,
	}
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

func (h Handler) GetAllTickets(c echo.Context) error {
	tickets, err := h.ticketRepository.GetAll(c.Request().Context())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, tickets)
}
