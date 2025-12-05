package events

type TicketRefunded struct {
	Header MessageHeader `json:"header"`
	TicketID string `json:"ticket_id"`
}

func NewTicketRefund (ticketId string) TicketRefunded {
	return TicketRefunded{
		Header: NewMessageHeader(),
		TicketID: ticketId,
	}
}