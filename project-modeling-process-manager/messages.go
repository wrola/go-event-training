package main

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type MessageHeader struct {
	ID             string    `json:"id"`
	PublishedAt    time.Time `json:"published_at"`
	IdempotencyKey string    `json:"idempotency_key"`
}

func NewMessageHeader() MessageHeader {
	return MessageHeader{
		ID:             uuid.NewString(),
		PublishedAt:    time.Now(),
		IdempotencyKey: uuid.NewString(),
	}
}

type Money struct {
	Amount   string `json:"amount" db:"amount"`
	Currency string `json:"currency" db:"currency"`
}

type BookShowTickets struct {
	BookingID uuid.UUID `json:"booking_id"`

	CustomerEmail   string    `json:"customer_email"`
	NumberOfTickets int       `json:"number_of_tickets"`
	ShowId          uuid.UUID `json:"show_id"`
}

type BookFlight struct {
	CustomerEmail  string    `json:"customer_email"`
	FlightID       uuid.UUID `json:"to_flight_id"`
	Passengers     []string  `json:"passengers"`
	ReferenceID    string    `json:"reference_id"`
	IdempotencyKey string    `json:"idempotency_key"`
}

type BookingMade_v1 struct {
	Header MessageHeader `json:"header"`

	NumberOfTickets int `json:"number_of_tickets"`

	BookingID uuid.UUID `json:"booking_id"`

	CustomerEmail string    `json:"customer_email"`
	ShowID        uuid.UUID `json:"show_id"`
}

type TicketBookingConfirmed_v1 struct {
	Header MessageHeader `json:"header"`

	TicketID      string `json:"ticket_id"`
	CustomerEmail string `json:"customer_email"`
	Price         Money  `json:"price"`

	BookingID string `json:"booking_id"`
}

type VipBundleInitialized_v1 struct {
	Header MessageHeader `json:"header"`

	VipBundleID VipBundleID `json:"vip_bundle_id"`
}

type FlightBooked_v1 struct {
	Header MessageHeader `json:"header"`

	FlightID  uuid.UUID   `json:"flight_id"`
	TicketIDs []uuid.UUID `json:"flight_tickets_ids"`

	ReferenceID string `json:"reference_id"`
}

type VipBundleFinalized_v1 struct {
	Header MessageHeader `json:"header"`

	VipBundleID VipBundleID `json:"vip_bundle_id"`
	Success     bool        `json:"success"`
}

type BookingFailed_v1 struct {
	Header MessageHeader `json:"header"`

	BookingID     uuid.UUID `json:"booking_id"`
	FailureReason string    `json:"failure_reason"`
}

type FlightBookingFailed_v1 struct {
	Header MessageHeader `json:"header"`

	FlightID      uuid.UUID `json:"flight_id"`
	FailureReason string    `json:"failure_reason"`

	ReferenceID string `json:"reference_id"`
}

type RefundTicket struct {
	Header MessageHeader `json:"header"`

	TicketID string `json:"ticket_id"`
}

type CommandBus interface {
	Send(ctx context.Context, command any) error
}

type EventBus interface {
	Publish(ctx context.Context, event any) error
}
