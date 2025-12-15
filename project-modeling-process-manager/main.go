package main

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type VipBundleID struct {
	uuid.UUID
}

func ParseBundleID(id string) (VipBundleID, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return VipBundleID{}, fmt.Errorf("failed to parse VipBundleID: %w", err)
	}
	return VipBundleID{UUID: parsed}, nil
}

func MustParseBundleID(id string) VipBundleID {
	parsed, err := uuid.Parse(id)
	if err != nil {
		panic(fmt.Sprintf("failed to parse VipBundleID: %s", id))
	}
	return VipBundleID{UUID: parsed}
}

type VipBundle struct {
	VipBundleID VipBundleID `json:"vip_bundle_id"`

	BookingID       uuid.UUID  `json:"booking_id"`
	CustomerEmail   string     `json:"customer_email"`
	NumberOfTickets int        `json:"number_of_tickets"`
	ShowId          uuid.UUID  `json:"show_id"`
	BookingMadeAt   *time.Time `json:"booking_made_at"`

	TicketIDs []uuid.UUID `json:"ticket_ids"`

	Passengers []string `json:"passengers"`

	InboundFlightID uuid.UUID `json:"inbound_flight_id"`

	IsFinalized bool `json:"is_finalized"`
	Failed 	bool `json:"failed"`
}

func NewVipBundle(
	vipBundleID VipBundleID,
	bookingID uuid.UUID,
	customerEmail string,
	numberOfTickets int,
	showId uuid.UUID,
	passengers []string,
	inboundFlightID uuid.UUID,
) (*VipBundle, error) {
	if vipBundleID.UUID == uuid.Nil {
		return nil, fmt.Errorf("vip bundle id must be set")
	}
	if bookingID == uuid.Nil {
		return nil, fmt.Errorf("booking id must be set")
	}
	if customerEmail == "" {
		return nil, fmt.Errorf("customer email must be set")
	}
	if numberOfTickets <= 0 {
		return nil, fmt.Errorf("number of tickets must be greater than 0")
	}
	if showId == uuid.Nil {
		return nil, fmt.Errorf("show id must be set")
	}
	if numberOfTickets != len(passengers) {
		return nil, fmt.Errorf("number of tickets and passengers count mismatch")
	}
	if inboundFlightID == uuid.Nil {
		return nil, fmt.Errorf("inbound flight id must be set")
	}

	return &VipBundle{
		VipBundleID:     vipBundleID,
		BookingID:       bookingID,
		CustomerEmail:   customerEmail,
		NumberOfTickets: numberOfTickets,
		ShowId:          showId,
		Passengers:      passengers,
		InboundFlightID: inboundFlightID,
	}, nil
}

type VipBundleRepository interface {
	Add(ctx context.Context, vipBundle VipBundle) error
	Get(ctx context.Context, vipBundleID VipBundleID) (VipBundle, error)
	GetByBookingID(ctx context.Context, bookingID uuid.UUID) (VipBundle, error)

	UpdateByID(
		ctx context.Context,
		vipBundleID VipBundleID,
		updateFn func(vipBundle VipBundle) (VipBundle, error),
	) (VipBundle, error)

	UpdateByBookingID(
		ctx context.Context,
		bookingID uuid.UUID,
		updateFn func(vipBundle VipBundle) (VipBundle, error),
	) (VipBundle, error)
}

type VipBundleProcessManager struct {
	commandBus CommandBus
	eventBus   EventBus
	repository VipBundleRepository
}

func NewVipBundleProcessManager(
	commandBus CommandBus,
	eventBus EventBus,
	repository VipBundleRepository,
) *VipBundleProcessManager {
	return &VipBundleProcessManager{
		commandBus: commandBus,
		eventBus:   eventBus,
		repository: repository,
	}
}

func (v VipBundleProcessManager) OnVipBundleInitialized(ctx context.Context, event *VipBundleInitialized_v1) error {
	vipBundle, err := v.repository.Get(ctx, event.VipBundleID)
	if err != nil {
		return fmt.Errorf("failed to get vip bundle %s: %w", event.VipBundleID, err)
	}

	cmd := BookShowTickets{
		BookingID:       vipBundle.BookingID,
		CustomerEmail:   vipBundle.CustomerEmail,
		NumberOfTickets: vipBundle.NumberOfTickets,
		ShowId:          vipBundle.ShowId,
	}

	err = v.commandBus.Send(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to send BookShowTickets command: %w", err)
	}

	return nil
}

func (v VipBundleProcessManager) OnBookingMade(ctx context.Context, event *BookingMade_v1) error {
	vipBundle, err := v.repository.UpdateByBookingID(
		ctx,
		event.BookingID,
		func(vipBundle VipBundle) (VipBundle, error) {
			now := time.Now()
			vipBundle.BookingMadeAt = &now
			return vipBundle, nil
		},
	)
	if err != nil {
		return fmt.Errorf("failed to update vip bundle for booking %s: %w", event.BookingID, err)
	}

	idempotencyKey := uuid.NewString()

	cmd := BookFlight{
		CustomerEmail:  vipBundle.CustomerEmail,
		FlightID:       vipBundle.InboundFlightID,
		Passengers:     vipBundle.Passengers,
		ReferenceID:    vipBundle.VipBundleID.String(),
		IdempotencyKey: idempotencyKey,
	}

	err = v.commandBus.Send(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to send BookFlight command: %w", err)
	}

	return nil
}

func (v VipBundleProcessManager) OnFlightBooked(ctx context.Context, event *FlightBooked_v1) error {
	vipBundleID := MustParseBundleID(event.ReferenceID)

	vipBundle, err := v.repository.UpdateByID(
		ctx,
		vipBundleID,
		func(vipBundle VipBundle) (VipBundle, error) {
			vipBundle.IsFinalized = true
			vipBundle.TicketIDs = event.TicketIDs
			return vipBundle, nil
		},
	)
	if err != nil {
		return fmt.Errorf("failed to update vip bundle %s: %w", vipBundleID, err)
	}

	finalizedEvent := VipBundleFinalized_v1{
		Header:      NewMessageHeader(),
		VipBundleID: vipBundle.VipBundleID,
		Success:     true,
	}

	err = v.eventBus.Publish(ctx, finalizedEvent)
	if err != nil {
		return fmt.Errorf("failed to publish VipBundleFinalized_v1 event: %w", err)
	}

	return nil
}

func (v VipBundleProcessManager) OnBookingFailed(ctx context.Context, event *BookingFailed_v1) error {
	vipBundle, err := v.repository.UpdateByBookingID(
		ctx,
		event.BookingID,
		func(vipBundle VipBundle) (VipBundle, error) {
			vipBundle.IsFinalized = true
			vipBundle.Failed = true
			return vipBundle, nil
		},
	)
	if err != nil {
		return fmt.Errorf("failed to update vip bundle for booking %s: %w", event.BookingID, err)
	}

	finalizedEvent := VipBundleFinalized_v1{
		Header:      NewMessageHeader(),
		VipBundleID: vipBundle.VipBundleID,
		Success:     false,
	}

	err = v.eventBus.Publish(ctx, finalizedEvent)
	if err != nil {
		return fmt.Errorf("failed to publish VipBundleFinalized_v1 event: %w", err)
	}

	return nil
}

func (v VipBundleProcessManager) OnTicketBookingConfirmed(ctx context.Context, event *TicketBookingConfirmed_v1) error {
	bookingID, err := uuid.Parse(event.BookingID)
	if err != nil {
		return fmt.Errorf("failed to parse booking ID %s: %w", event.BookingID, err)
	}

	ticketID, err := uuid.Parse(event.TicketID)
	if err != nil {
		return fmt.Errorf("failed to parse ticket ID %s: %w", event.TicketID, err)
	}

	_, err = v.repository.UpdateByBookingID(
		ctx,
		bookingID,
		func(vipBundle VipBundle) (VipBundle, error) {
			vipBundle.TicketIDs = append(vipBundle.TicketIDs, ticketID)
			return vipBundle, nil
		},
	)
	if err != nil {
		return fmt.Errorf("failed to update vip bundle for booking %s: %w", bookingID, err)
	}

	return nil
}

func (v VipBundleProcessManager) OnFlightBookingFailed(ctx context.Context, event *FlightBookingFailed_v1) error {
	vipBundleID := MustParseBundleID(event.ReferenceID)

	vipBundle, err := v.repository.Get(ctx, vipBundleID)
	if err != nil {
		return fmt.Errorf("failed to get vip bundle %s: %w", vipBundleID, err)
	}

	if vipBundle.BookingMadeAt == nil {
		return nil
	}

	if len(vipBundle.TicketIDs) != vipBundle.NumberOfTickets {
		return fmt.Errorf(
			"not all tickets confirmed yet: expected %d, got %d",
			vipBundle.NumberOfTickets,
			len(vipBundle.TicketIDs),
		)
	}

	for _, ticketID := range vipBundle.TicketIDs {
		cmd := RefundTicket{
			Header:   NewMessageHeader(),
			TicketID: ticketID.String(),
		}

		err = v.commandBus.Send(ctx, cmd)
		if err != nil {
			return fmt.Errorf("failed to send RefundTicket command for ticket %s: %w", ticketID, err)
		}
	}

	vipBundle, err = v.repository.UpdateByID(
		ctx,
		vipBundleID,
		func(vipBundle VipBundle) (VipBundle, error) {
			vipBundle.IsFinalized = true
			vipBundle.Failed = true
			return vipBundle, nil
		},
	)
	if err != nil {
		return fmt.Errorf("failed to update vip bundle %s: %w", vipBundleID, err)
	}

	finalizedEvent := VipBundleFinalized_v1{
		Header:      NewMessageHeader(),
		VipBundleID: vipBundle.VipBundleID,
		Success:     false,
	}

	err = v.eventBus.Publish(ctx, finalizedEvent)
	if err != nil {
		return fmt.Errorf("failed to publish VipBundleFinalized_v1 event: %w", err)
	}

	return nil
}
