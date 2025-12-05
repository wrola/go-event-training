package events
import (
	"time"
)

type TicketReceiptIssued struct {
	Header MessageHeader `json:"header"`

	TicketID      string `json:"ticket_id"`
	ReceiptNumber string `json:"receipt_number"`

	IssuedAt time.Time `json:"issued_at"`
}

func NewTicketReceiptIssued (ticketID string, receiptNumber string, issuedAt time.Time,  idempotencyKey string) TicketReceiptIssued {

	return TicketReceiptIssued{
		TicketID: ticketID,
		ReceiptNumber: receiptNumber,
		IssuedAt: issuedAt,
		Header: NewMessageHeaderWithIdempotencyKey(idempotencyKey),
	}
}