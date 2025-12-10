package events

type InternalOpsReadModelUpdated struct {
	Header    MessageHeader `json:"header"`
	BookingID string        `json:"booking_id"`
}

func NewInternalOpsReadModelUpdated(bookingID string) InternalOpsReadModelUpdated {
	return InternalOpsReadModelUpdated{
		Header:    NewMessageHeader(),
		BookingID: bookingID,
	}
}

func (e InternalOpsReadModelUpdated) IsInternal() bool {
	return true
}
