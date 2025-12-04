package events

import "tickets/entities/models"

type TicketBookingCanceled struct {
	Header        MessageHeader `json:"header"`
	TicketID      string        `json:"ticket_id"`
	CustomerEmail string        `json:"customer_email"`
	Price         models.Money  `json:"price"`
}

func NewTicketBookingCanceled(ticketID, customerEmail string, price models.Money) TicketBookingCanceled {
	return TicketBookingCanceled{
		Header:        NewMessageHeader(),
		TicketID:      ticketID,
		CustomerEmail: customerEmail,
		Price:         price,
	}
}

func NewTicketBookingCanceledWithIdempotencyKey(ticketID, customerEmail string, price models.Money, idempotencyKey string) TicketBookingCanceled {
	return TicketBookingCanceled{
		Header:        NewMessageHeaderWithIdempotencyKey(idempotencyKey),
		TicketID:      ticketID,
		CustomerEmail: customerEmail,
		Price:         price,
	}
}
