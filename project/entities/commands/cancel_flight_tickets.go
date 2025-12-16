package commands

type CancelFlightTickets struct {
	Header         MessageHeader `json:"header"`
	FlightTicketIDs []string      `json:"flight_ticket_ids"`
}

func NewCancelFlightTickets(flightTicketIDs []string) CancelFlightTickets {
	return CancelFlightTickets{
		Header:          NewMessageHeader(),
		FlightTicketIDs: flightTicketIDs,
	}
}
