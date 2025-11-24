package entities

type Ticket struct {
	TicketID      string `json:"ticket_id"`
	Status        string `json:"status"`
	CustomerEmail string `json:"customer_email"`
	Price         Money  `json:"price"`
}

type TicketBookingStatus string

const (
	TicketStatusConfirmed TicketBookingStatus = "confirmed"
	TicketStatusCanceled  TicketBookingStatus = "canceled"
)
