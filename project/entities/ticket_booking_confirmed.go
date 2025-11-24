package entities

type TicketBookingConfirmed struct {
	Header        MessageHeader `json:"header"`
	TicketID      string        `json:"ticket_id"`
	CustomerEmail string        `json:"customer_email"`
	Price         Money         `json:"price"`
}

func NewTicketBookingConfirmed(ticketID, customerEmail string, price Money) TicketBookingConfirmed {
	return TicketBookingConfirmed{
		Header:        NewMessageHeader(),
		TicketID:      ticketID,
		CustomerEmail: customerEmail,
		Price:         price,
	}
}
