package events

import "tickets/entities/models"

type TicketBookingConfirmed struct {
	Header        MessageHeader `json:"header"`
	TicketID      string        `json:"ticket_id"`
	CustomerEmail string        `json:"customer_email"`
	Price         models.Money  `json:"price"`
}

func NewTicketBookingConfirmed(ticketID, customerEmail string, price models.Money) TicketBookingConfirmed {
	return TicketBookingConfirmed{
		Header:        NewMessageHeader(),
		TicketID:      ticketID,
		CustomerEmail: customerEmail,
		Price:         price,
	}
}

func NewTicketBookingConfirmedWithIdempotencyKey(ticketID, customerEmail string, price models.Money, idempotencyKey string) TicketBookingConfirmed {
	return TicketBookingConfirmed{
		Header:        NewMessageHeaderWithIdempotencyKey(idempotencyKey),
		TicketID:      ticketID,
		CustomerEmail: customerEmail,
		Price:         price,
	}
}
