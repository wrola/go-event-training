package commands

type BookTaxi struct {
	Header             MessageHeader `json:"header"`
	CustomerEmail      string        `json:"customer_email"`
	CustomerName       string        `json:"customer_name"`
	NumberOfPassengers int           `json:"number_of_passengers"`
	ReferenceID        string        `json:"reference_id"`
	IdempotencyKey     string        `json:"idempotency_key"`
}

func NewBookTaxi(customerEmail, customerName string, numberOfPassengers int, referenceID, idempotencyKey string) BookTaxi {
	return BookTaxi{
		Header:             NewMessageHeader(),
		CustomerEmail:      customerEmail,
		CustomerName:       customerName,
		NumberOfPassengers: numberOfPassengers,
		ReferenceID:        referenceID,
		IdempotencyKey:     idempotencyKey,
	}
}
