package events

type BookingMade struct {
    Header MessageHeader `json:"header"`

    NumberOfTickets int    `json:"number_of_tickets"`
    BookingID       string `json:"booking_id"`
    CustomerEmail   string `json:"customer_email"`
    ShowID          string `json:"show_id"`
}

func NewBookingMade(bookingID string, customerEmail string, showID string, numberOfTickets int, idempotencyKey string) BookingMade {
    return BookingMade{
        Header: NewMessageHeaderWithIdempotencyKey(idempotencyKey),
        BookingID: bookingID,
        CustomerEmail: customerEmail,
        ShowID: showID,
        NumberOfTickets: numberOfTickets,
    }
}
