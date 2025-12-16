package commands

type BookShowTickets struct {
	Header          MessageHeader `json:"header"`
	BookingID       string        `json:"booking_id"`
	CustomerEmail   string        `json:"customer_email"`
	NumberOfTickets int           `json:"number_of_tickets"`
	ShowID          string        `json:"show_id"`
}

func NewBookShowTickets(bookingID string, customerEmail string, numberOfTickets int, showID string) BookShowTickets {
	return BookShowTickets{
		Header:          NewMessageHeader(),
		BookingID:       bookingID,
		CustomerEmail:   customerEmail,
		NumberOfTickets: numberOfTickets,
		ShowID:          showID,
	}
}
