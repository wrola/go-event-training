package models

import (
	"time"

	"github.com/google/uuid"
)

type OpsBooking struct {
	BookingID  uuid.UUID            `json:"booking_id"`
	BookedAt   time.Time            `json:"booked_at"`
	Tickets    map[string]OpsTicket `json:"tickets"`
	LastUpdate time.Time            `json:"last_update"`
}

type OpsTicket struct {
	PriceAmount     string    `json:"price_amount"`
	PriceCurrency   string    `json:"price_currency"`
	CustomerEmail   string    `json:"customer_email"`
	Status          string    `json:"status"`
	PrintedAt       time.Time `json:"printed_at,omitempty"`
	PrintedFileName string    `json:"printed_file_name,omitempty"`
	ReceiptIssuedAt time.Time `json:"receipt_issued_at,omitempty"`
	ReceiptNumber   string    `json:"receipt_number,omitempty"`
}
