package events

import "tickets/entities/models"

type TicketBookingCanceled_v1 struct {
	Header        MessageHeader `json:"header"`
	TicketID      string        `json:"ticket_id"`
	CustomerEmail string        `json:"customer_email"`
	Price         models.Money  `json:"price"`
}

func NewTicketBookingCanceled_v1(ticketID, customerEmail string, price models.Money) TicketBookingCanceled_v1 {
	return TicketBookingCanceled_v1{
		Header:        NewMessageHeader(),
		TicketID:      ticketID,
		CustomerEmail: customerEmail,
		Price:         price,
	}
}

func NewTicketBookingCanceled_v1WithIdempotencyKey(ticketID, customerEmail string, price models.Money, idempotencyKey string) TicketBookingCanceled_v1 {
	return TicketBookingCanceled_v1{
		Header:        NewMessageHeaderWithIdempotencyKey(idempotencyKey),
		TicketID:      ticketID,
		CustomerEmail: customerEmail,
		Price:         price,
	}
}
