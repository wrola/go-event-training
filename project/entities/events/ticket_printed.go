package events

type TicketPrinted struct {
	Header MessageHeader `json:"header"`

	TicketID string `json:"ticket_id"`
	FileName string `json:"file_name"`
}

func NewTicketPrinted(ticketID, fileName string) TicketPrinted {
	return TicketPrinted{
		Header:   NewMessageHeader(),
		TicketID: ticketID,
		FileName: fileName,
	}
}
