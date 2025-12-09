package http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"tickets/database"
	"tickets/entities/commands"
	"tickets/entities/events"
	"tickets/entities/models"
	"github.com/google/uuid"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
)

type Handler struct {
	eventBus           *cqrs.EventBus
	commandBus         *cqrs.CommandBus
	ticketRepository   database.TicketRepository
	showsRepository    database.ShowsRepository
	bookingsRepository database.BookingsRepository
	opsBookingRepository database.OpsBookingReadModelRepository
}

func NewHandler(eventBus *cqrs.EventBus, commandBus *cqrs.CommandBus, ticketRepository database.TicketRepository, showsRepository database.ShowsRepository, bookingsRepository database.BookingsRepository, opsBookingRepository database.OpsBookingReadModelRepository) Handler {
	if eventBus == nil {
		panic("eventBus is required")
	}
	if commandBus == nil {
		panic("commandBus is required")
	}

	return Handler{
		eventBus:             eventBus,
		commandBus:           commandBus,
		ticketRepository:     ticketRepository,
		showsRepository:      showsRepository,
		bookingsRepository:   bookingsRepository,
		opsBookingRepository: opsBookingRepository,
	}
}

type TicketStatusRequest struct {
	TicketID string `json:"ticket_id"`
	CustomerEmail string `json:"customer_email"`
	Price models.Money `json:"price"`
	Status models.TicketBookingStatus `json:"status"`
	BookingID string `json:"booking_id"`
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
	BookingID		string  `json:"booking_id"`
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
  			event := events.NewTicketBookingConfirmed_v1WithIdempotencyKey(
  				ticket.TicketID,
  				ticket.CustomerEmail,
  				ticket.Price,
  				ticketIdempotencyKey,
				ticket.BookingID,
  			)
  			h.eventBus.Publish(c.Request().Context(), event)
  		case models.TicketStatusCanceled:
  			ticketIdempotencyKey := idempotencyKey + ticket.TicketID
  			event := events.NewTicketBookingCanceled_v1WithIdempotencyKey(
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
		request.BookingID,
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

func (h Handler) PutTicketRefund(c echo.Context) error {
	ticketID := c.Param("ticket_id")

	if ticketID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "ticket_id is required")
	}

	idempotencyKey := c.Request().Header.Get("Idempotency-Key")

	command := commands.NewRefundTicketWithIdempotencyKey(ticketID, idempotencyKey)

	if err := h.commandBus.Send(c.Request().Context(), command); err != nil {
		return err
	}

	return c.NoContent(http.StatusAccepted)
}

func (h Handler) GetOpsBookings(c echo.Context) error {
	// Parse optional query parameter
	receiptIssueDateStr := c.QueryParam("receipt_issue_date")

	var receiptIssueDate *time.Time
	if receiptIssueDateStr != "" {
		parsed, err := time.Parse("2006-01-02", receiptIssueDateStr)
		if err != nil {
			return echo.NewHTTPError(
				http.StatusBadRequest,
				fmt.Sprintf("invalid date format for receipt_issue_date: expected YYYY-MM-DD, got %q", receiptIssueDateStr),
			)
		}
		receiptIssueDate = &parsed
	}

	opsBookings, err := h.opsBookingRepository.GetAllReadModels(
		c.Request().Context(),
		receiptIssueDate,
	)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, opsBookings)
}

func (h Handler) GetOpsBookingByID(c echo.Context) error {
	bookingID := c.Param("booking_id")

	if bookingID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "booking_id is required")
	} 

	if err := uuid.Validate(bookingID); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "booking_id is not a valid UUID")
	}

	opsBooking, err := h.opsBookingRepository.FindReadModelByBookingID(c.Request().Context(), bookingID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, opsBooking)
}