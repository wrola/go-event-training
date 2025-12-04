package commands

type RefundTicket struct {
	Header   MessageHeader `json:"header"`
	TicketID string        `json:"ticket_id"`
}

func NewRefundTicket(ticketID string) RefundTicket {
	return RefundTicket{
		Header:   NewMessageHeader(),
		TicketID: ticketID,
	}
}

func NewRefundTicketWithIdempotencyKey(ticketID string, idempotencyKey string) RefundTicket {
	return RefundTicket{
		Header:   NewMessageHeaderWithIdempotencyKey(idempotencyKey),
		TicketID: ticketID,
	}
}
