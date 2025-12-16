package commands

type BookFlight struct {
	Header         MessageHeader `json:"header"`
	CustomerEmail  string        `json:"customer_email"`
	FlightID       string        `json:"flight_id"`
	Passengers     []string      `json:"passengers"`
	ReferenceID    string        `json:"reference_id"`
	IdempotencyKey string        `json:"idempotency_key"`
}

func NewBookFlight(customerEmail string, flightID string, passengers []string, referenceID string, idempotencyKey string) BookFlight {
	return BookFlight{
		Header:         NewMessageHeader(),
		CustomerEmail:  customerEmail,
		FlightID:       flightID,
		Passengers:     passengers,
		ReferenceID:    referenceID,
		IdempotencyKey: idempotencyKey,
	}
}
