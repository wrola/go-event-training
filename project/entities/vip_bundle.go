package entities

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type VipBundleID struct {
	uuid.UUID
}

func NewVipBundleID() VipBundleID {
	return VipBundleID{UUID: uuid.New()}
}

func ParseVipBundleID(id string) (VipBundleID, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return VipBundleID{}, fmt.Errorf("failed to parse VipBundleID: %w", err)
	}
	return VipBundleID{UUID: parsed}, nil
}

func MustParseVipBundleID(id string) VipBundleID {
	parsed, err := uuid.Parse(id)
	if err != nil {
		panic(fmt.Sprintf("failed to parse VipBundleID: %s", id))
	}
	return VipBundleID{UUID: parsed}
}

type VipBundle struct {
	VipBundleID     VipBundleID `json:"vip_bundle_id"`
	BookingID       string      `json:"booking_id"`
	CustomerEmail   string      `json:"customer_email"`
	NumberOfTickets int         `json:"number_of_tickets"`
	ShowID          string      `json:"show_id"`
	BookingMadeAt   *time.Time  `json:"booking_made_at"`
	TicketIDs       []string    `json:"ticket_ids"`
	Passengers      []string    `json:"passengers"`
	InboundFlightID string      `json:"inbound_flight_id"`
	IsFinalized     bool        `json:"is_finalized"`
	Failed          bool        `json:"failed"`
}

func NewVipBundle(
	vipBundleID VipBundleID,
	bookingID string,
	customerEmail string,
	numberOfTickets int,
	showID string,
	passengers []string,
	inboundFlightID string,
) (*VipBundle, error) {
	if vipBundleID.UUID == uuid.Nil {
		return nil, fmt.Errorf("vip bundle id must be set")
	}
	if bookingID == "" {
		return nil, fmt.Errorf("booking id must be set")
	}
	if customerEmail == "" {
		return nil, fmt.Errorf("customer email must be set")
	}
	if numberOfTickets <= 0 {
		return nil, fmt.Errorf("number of tickets must be greater than 0")
	}
	if showID == "" {
		return nil, fmt.Errorf("show id must be set")
	}
	if len(passengers) == 0 {
		return nil, fmt.Errorf("at least one passenger must be specified")
	}
	if inboundFlightID == "" {
		return nil, fmt.Errorf("inbound flight id must be set")
	}

	return &VipBundle{
		VipBundleID:     vipBundleID,
		BookingID:       bookingID,
		CustomerEmail:   customerEmail,
		NumberOfTickets: numberOfTickets,
		ShowID:          showID,
		Passengers:      passengers,
		InboundFlightID: inboundFlightID,
		TicketIDs:       []string{},
		BookingMadeAt:   nil,
		IsFinalized:     false,
		Failed:          false,
	}, nil
}
