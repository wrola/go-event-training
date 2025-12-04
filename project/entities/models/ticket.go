package models

type Ticket struct {
	TicketID      string `json:"ticket_id" db:"ticket_id"`
	Status        string `json:"status" db:"status"`
	CustomerEmail string `json:"customer_email" db:"customer_email"`
	Price         Money  `json:"price" db:"price"`
}

type TicketBookingStatus string

const (
	TicketStatusConfirmed TicketBookingStatus = "confirmed"
	TicketStatusCanceled  TicketBookingStatus = "canceled"
)
