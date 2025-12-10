package events

import "tickets/entities/models"

type TicketBookingConfirmed_v1 struct {
	Header        MessageHeader `json:"header"`
	TicketID      string        `json:"ticket_id"`
	CustomerEmail string        `json:"customer_email"`
	Price         models.Money  `json:"price"`
	BookingID 	string 			`json:"booking_id"`
}

func NewTicketBookingConfirmed_v1(ticketID, customerEmail string, price models.Money, bookingID string) TicketBookingConfirmed_v1 {
	return TicketBookingConfirmed_v1{
		Header:        NewMessageHeader(),
		TicketID:      ticketID,
		CustomerEmail: customerEmail,
		Price:         price,
		BookingID:     bookingID,
	}
}

func NewTicketBookingConfirmed_v1WithIdempotencyKey(ticketID, customerEmail string, price models.Money, idempotencyKey string, bookingID string) TicketBookingConfirmed_v1 {
	return TicketBookingConfirmed_v1{
		Header:        NewMessageHeaderWithIdempotencyKey(idempotencyKey),
		TicketID:      ticketID,
		CustomerEmail: customerEmail,
		Price:         price,
		BookingID: bookingID,
	}
}

func (e TicketBookingConfirmed_v1) IsInternal() bool {
	return false
}
