package events
import (
	"time"
)

type TicketReceiptIssued_v1 struct {
	Header MessageHeader `json:"header"`

	TicketID      string `json:"ticket_id"`
	ReceiptNumber string `json:"receipt_number"`

	IssuedAt time.Time `json:"issued_at"`
}

func NewTicketReceiptIssued_v1 (ticketID string, receiptNumber string, issuedAt time.Time,  idempotencyKey string) TicketReceiptIssued_v1 {

	return TicketReceiptIssued_v1{
		TicketID: ticketID,
		ReceiptNumber: receiptNumber,
		IssuedAt: issuedAt,
		Header: NewMessageHeaderWithIdempotencyKey(idempotencyKey),
	}
}