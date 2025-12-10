package events

type TicketRefunded_v1 struct {
	Header MessageHeader `json:"header"`
	TicketID string `json:"ticket_id"`
}

func NewTicketRefund (ticketId string) TicketRefunded_v1 {
	return TicketRefunded_v1{
		Header: NewMessageHeader(),
		TicketID: ticketId,
	}
}

func (e TicketRefunded_v1) IsInternal() bool {
	return false
}
