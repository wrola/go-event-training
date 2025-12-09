package models

type TicketBookingStatus string

const (
	TicketStatusConfirmed TicketBookingStatus = "confirmed"
	TicketStatusCanceled  TicketBookingStatus = "canceled"
)

type Ticket struct {
	TicketID      string `json:"ticket_id" db:"ticket_id"`
	CustomerEmail string `json:"customer_email" db:"customer_email"`
	Price         Money  `json:"price" db:"price"`
	ConfirmedAt   string `json:"-" db:"confirmed_at"`
	RefundedAt    string `json:"-" db:"refunded_at"`
	DeletedAt     string `json:"-" db:"deleted_at"`
}
