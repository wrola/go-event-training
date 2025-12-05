package commands

type PaymentRefund struct {
	TicketID       string
	RefundReason   string
	IdempotencyKey string
}

