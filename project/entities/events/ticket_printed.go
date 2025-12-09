package events

type TicketPrinted_v1 struct {
	Header MessageHeader `json:"header"`

	TicketID string `json:"ticket_id"`
	FileName string `json:"file_name"`
}

func NewTicketPrinted_v1(ticketID, fileName string) TicketPrinted_v1 {
	return TicketPrinted_v1{
		Header:   NewMessageHeader(),
		TicketID: ticketID,
		FileName: fileName,
	}
}
