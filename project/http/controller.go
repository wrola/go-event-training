package http

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"tickets/database"
	"tickets/entities/models"
	"tickets/entities/events"
)

type Handler struct {
	eventBus           *cqrs.EventBus
	ticketRepository   database.TicketRepository
	showsRepository    database.ShowsRepository
	bookingsRepository database.BookingsRepository
}

func NewHandler(eventBus *cqrs.EventBus, ticketRepository database.TicketRepository, showsRepository database.ShowsRepository, bookingsRepository database.BookingsRepository) Handler {
	if eventBus == nil {
		panic("eventBus is required")
	}

	return Handler{
		eventBus:           eventBus,
		ticketRepository:   ticketRepository,
		showsRepository:    showsRepository,
		bookingsRepository: bookingsRepository,
	}
}

type TicketStatusRequest struct {
	TicketID string `json:"ticket_id"`
	CustomerEmail string `json:"customer_email"`
	Price models.Money `json:"price"`
	Status models.TicketBookingStatus `json:"status"`
}

type ticketsConfirmationRequest struct {
	Tickets []TicketStatusRequest `json:"tickets"`
}

type ShowCreationRequest struct {
	ShowID          string    `json:"show_id"`
	DeadNationID    string    `json:"dead_nation_id"`
	NumberOfTickets int       `json:"number_of_tickets"`
	StartTime       time.Time `json:"start_time"`
	Title           string    `json:"title"`
	Venue           string    `json:"venue"`
}

type BookTicketsRequest struct {
	ShowID          string `json:"show_id"`
	NumberOfTickets int    `json:"number_of_tickets"`
	CustomerEmail   string `json:"customer_email"`
}

type BookTicketsResponse struct {
	BookingID string `json:"booking_id"`
}

func (h Handler) PostTicketsConfirmation(c echo.Context) error {
	var request ticketsConfirmationRequest
	if err := c.Bind(&request);  err != nil {
		return err
	}

	idempotencyKey := c.Request().Header.Get("Idempotency-Key")
	if idempotencyKey == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Idempotency-Key header is required")
	}

	for _, ticket := range request.Tickets {
  	switch ticket.Status {
  		case models.TicketStatusConfirmed:
  			ticketIdempotencyKey := idempotencyKey + ticket.TicketID
  			event := events.NewTicketBookingConfirmedWithIdempotencyKey(
  				ticket.TicketID,
  				ticket.CustomerEmail,
  				ticket.Price,
  				ticketIdempotencyKey,
  			)
  			h.eventBus.Publish(c.Request().Context(), event)
  		case models.TicketStatusCanceled:
  			ticketIdempotencyKey := idempotencyKey + ticket.TicketID
  			event := events.NewTicketBookingCanceledWithIdempotencyKey(
  				ticket.TicketID,
  				ticket.CustomerEmail,
  				ticket.Price,
  				ticketIdempotencyKey,
  			)
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

func (h Handler) CreateShow(c echo.Context) error {
	var request ShowCreationRequest
	if err := c.Bind(&request);  err != nil {
		return err
	}

	// idempotencyKey := c.Request().Header.Get("Idempotency-Key")
	// if idempotencyKey == "" {
	// 	return echo.NewHTTPError(http.StatusBadRequest, "Idempotency-Key header is required")
	// }

	show := models.NewShow(
		request.ShowID,
		request.DeadNationID,
		request.NumberOfTickets,
		request.StartTime,
		request.Title,
		request.Venue,
	)

	err := h.showsRepository.AddShow(c.Request().Context(), show)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, show)
}

func (h Handler) BookTickets(c echo.Context) error {
	var request BookTicketsRequest
	if err := c.Bind(&request); err != nil {
		return err
	}

	booking := models.NewBooking(
		"", // booking_id will be auto-generated
		request.ShowID,
		request.NumberOfTickets,
		request.CustomerEmail,
	)

	err := h.bookingsRepository.AddBooking(c.Request().Context(), booking)
	if err != nil {
		return err
	}

	response := BookTicketsResponse{
		BookingID: booking.BookingID,
	}

	return c.JSON(http.StatusCreated, response)
}