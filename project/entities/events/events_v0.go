package events

import (
	"time"

	"github.com/google/uuid"

	"tickets/entities/models"
)

type BookingMade_v0 struct {
	Header MessageHeader `json:"header"`

	NumberOfTickets int `json:"number_of_tickets"`

	BookingID uuid.UUID `json:"booking_id"`

	CustomerEmail string    `json:"customer_email"`
	ShowID        uuid.UUID `json:"show_id"`
}

type TicketBookingConfirmed_v0 struct {
	Header MessageHeader `json:"header"`

	TicketID      string       `json:"ticket_id"`
	CustomerEmail string       `json:"customer_email"`
	Price         models.Money `json:"price"`

	BookingID string `json:"booking_id"`
}

type TicketReceiptIssued_v0 struct {
	Header MessageHeader `json:"header"`

	TicketID      string `json:"ticket_id"`
	ReceiptNumber string `json:"receipt_number"`

	IssuedAt time.Time `json:"issued_at"`
}

type TicketPrinted_v0 struct {
	Header MessageHeader `json:"header"`

	TicketID string `json:"ticket_id"`
	FileName string `json:"file_name"`
}

type TicketRefunded_v0 struct {
	Header MessageHeader `json:"header"`

	TicketID string `json:"ticket_id"`
}
