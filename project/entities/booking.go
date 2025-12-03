package entities

import ( 
	"github.com/google/uuid"
)

type Booking struct {
    BookingID       string `json:"booking_id" db:"booking_id"`
    ShowID          string `json:"show_id" db:"show_id"`
    NumberOfTickets int    `json:"number_of_tickets" db:"number_of_tickets"`
    CustomerEmail   string `json:"customer_email" db:"customer_email"`
}

func NewBooking(bookingID string, showID string, numberOfTickets int, customerEmail string) Booking {

	if bookingID == "" {
		bookingID = uuid.NewString()
	}
	
	return Booking{
		BookingID: bookingID,
		ShowID: showID,
		NumberOfTickets: numberOfTickets,
		CustomerEmail: customerEmail,
	}

}