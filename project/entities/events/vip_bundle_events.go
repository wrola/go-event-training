package events

import "tickets/entities"

type VipBundleInitialized_v1 struct {
	Header      MessageHeader       `json:"header"`
	VipBundleID entities.VipBundleID `json:"vip_bundle_id"`
}

func NewVipBundleInitialized(vipBundleID entities.VipBundleID) VipBundleInitialized_v1 {
	return VipBundleInitialized_v1{
		Header:      NewMessageHeader(),
		VipBundleID: vipBundleID,
	}
}

func (e VipBundleInitialized_v1) IsInternal() bool {
	return false
}

type BookingFailed_v1 struct {
	Header        MessageHeader `json:"header"`
	BookingID     string        `json:"booking_id"`
	FailureReason string        `json:"failure_reason"`
}

func NewBookingFailed(bookingID string, reason string) BookingFailed_v1 {
	return BookingFailed_v1{
		Header:        NewMessageHeader(),
		BookingID:     bookingID,
		FailureReason: reason,
	}
}

func (e BookingFailed_v1) IsInternal() bool {
	return false
}

type FlightBooked_v1 struct {
	Header      MessageHeader `json:"header"`
	FlightID    string        `json:"flight_id"`
	TicketIDs   []string      `json:"flight_tickets_ids"`
	ReferenceID string        `json:"reference_id"`
}

func NewFlightBooked(flightID string, ticketIDs []string, referenceID string) FlightBooked_v1 {
	return FlightBooked_v1{
		Header:      NewMessageHeader(),
		FlightID:    flightID,
		TicketIDs:   ticketIDs,
		ReferenceID: referenceID,
	}
}

func (e FlightBooked_v1) IsInternal() bool {
	return false
}

type FlightBookingFailed_v1 struct {
	Header        MessageHeader `json:"header"`
	FlightID      string        `json:"flight_id"`
	FailureReason string        `json:"failure_reason"`
	ReferenceID   string        `json:"reference_id"`
}

func NewFlightBookingFailed(flightID string, reason string, referenceID string) FlightBookingFailed_v1 {
	return FlightBookingFailed_v1{
		Header:        NewMessageHeader(),
		FlightID:      flightID,
		FailureReason: reason,
		ReferenceID:   referenceID,
	}
}

func (e FlightBookingFailed_v1) IsInternal() bool {
	return false
}

type VipBundleFinalized_v1 struct {
	Header      MessageHeader       `json:"header"`
	VipBundleID entities.VipBundleID `json:"vip_bundle_id"`
	Success     bool                `json:"success"`
}

func NewVipBundleFinalized(vipBundleID entities.VipBundleID, success bool) VipBundleFinalized_v1 {
	return VipBundleFinalized_v1{
		Header:      NewMessageHeader(),
		VipBundleID: vipBundleID,
		Success:     success,
	}
}

func (e VipBundleFinalized_v1) IsInternal() bool {
	return false
}
type TaxiBooked_v1 struct {
	Header        MessageHeader `json:"header"`
	TaxiBookingID string        `json:"taxi_booking_id"`
	ReferenceID   string        `json:"reference_id"`
}

func NewTaxiBooked(taxiBookingID string, referenceID string) TaxiBooked_v1 {
	return TaxiBooked_v1{
		Header:        NewMessageHeader(),
		TaxiBookingID: taxiBookingID,
		ReferenceID:   referenceID,
	}
}

func (e TaxiBooked_v1) IsInternal() bool {
	return false
}

type TaxiBookingFailed_v1 struct {
	Header        MessageHeader `json:"header"`
	FailureReason string        `json:"failure_reason"`
	ReferenceID   string        `json:"reference_id"`
}

func NewTaxiBookingFailed(reason string, referenceID string) TaxiBookingFailed_v1 {
	return TaxiBookingFailed_v1{
		Header:        NewMessageHeader(),
		FailureReason: reason,
		ReferenceID:   referenceID,
	}
}

func (e TaxiBookingFailed_v1) IsInternal() bool {
	return false
}
