package process_manager

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/google/uuid"

	"tickets/database"
	"tickets/entities"
	"tickets/entities/commands"
	"tickets/entities/events"
)

type VipBundleProcessManager struct {
	commandBus *cqrs.CommandBus
	eventBus   *cqrs.EventBus
	repository database.VipBundleRepository
}

func NewVipBundleProcessManager(
	commandBus *cqrs.CommandBus,
	eventBus *cqrs.EventBus,
	repository database.VipBundleRepository,
) *VipBundleProcessManager {
	if commandBus == nil {
		panic("commandBus is required")
	}
	if eventBus == nil {
		panic("eventBus is required")
	}

	return &VipBundleProcessManager{
		commandBus: commandBus,
		eventBus:   eventBus,
		repository: repository,
	}
}

func (pm *VipBundleProcessManager) OnVipBundleInitialized(ctx context.Context, event *events.VipBundleInitialized_v1) error {
	vipBundle, err := pm.repository.Get(ctx, event.VipBundleID)
	if err != nil {
		return fmt.Errorf("failed to get vip bundle %s: %w", event.VipBundleID, err)
	}

	cmd := commands.NewBookShowTickets(
		vipBundle.BookingID,
		vipBundle.CustomerEmail,
		vipBundle.NumberOfTickets,
		vipBundle.ShowID,
	)

	err = pm.commandBus.Send(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to send BookShowTickets command: %w", err)
	}

	return nil
}

func (pm *VipBundleProcessManager) OnBookingMade(ctx context.Context, event *events.BookingMade_v1) error {
	vipBundle, err := pm.repository.UpdateByBookingID(
		ctx,
		event.BookingID,
		func(vipBundle entities.VipBundle) (entities.VipBundle, error) {
			now := time.Now()
			vipBundle.BookingMadeAt = &now
			return vipBundle, nil
		},
	)
	if err != nil {
		if err.Error() == "sql: no rows in result set" || err == sql.ErrNoRows {
			return nil
		}
		return fmt.Errorf("failed to update vip bundle for booking %s: %w", event.BookingID, err)
	}

	idempotencyKey := uuid.NewString()
	cmd := commands.NewBookFlight(
		vipBundle.CustomerEmail,
		vipBundle.InboundFlightID,
		vipBundle.Passengers,
		vipBundle.VipBundleID.String(),
		idempotencyKey,
	)
	err = pm.commandBus.Send(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to send BookFlight command: %w", err)
	}

	return nil
}

func (pm *VipBundleProcessManager) OnTicketBookingConfirmed(ctx context.Context, event *events.TicketBookingConfirmed_v1) error {
	_, err := pm.repository.UpdateByBookingID(
		ctx,
		event.BookingID,
		func(vipBundle entities.VipBundle) (entities.VipBundle, error) {
			vipBundle.TicketIDs = append(vipBundle.TicketIDs, event.TicketID)
			return vipBundle, nil
		},
	)
	if err != nil {
		// Not all ticket confirmations are part of VIP bundles - ignore if not found
		if err.Error() == "sql: no rows in result set" || err == sql.ErrNoRows {
			return nil
		}
		return fmt.Errorf("failed to update vip bundle for booking %s: %w", event.BookingID, err)
	}

	return nil
}

func (pm *VipBundleProcessManager) OnBookingFailed(ctx context.Context, event *events.BookingFailed_v1) error {
	vipBundle, err := pm.repository.UpdateByBookingID(
		ctx,
		event.BookingID,
		func(vipBundle entities.VipBundle) (entities.VipBundle, error) {
			vipBundle.IsFinalized = true
			vipBundle.Failed = true
			return vipBundle, nil
		},
	)
	if err != nil {
		if err.Error() == "sql: no rows in result set" || err == sql.ErrNoRows {
			return nil
		}
		return fmt.Errorf("failed to update vip bundle for booking %s: %w", event.BookingID, err)
	}

	finalizedEvent := events.NewVipBundleFinalized(vipBundle.VipBundleID, false)
	err = pm.eventBus.Publish(ctx, finalizedEvent)
	if err != nil {
		return fmt.Errorf("failed to publish VipBundleFinalized_v1 event: %w", err)
	}

	return nil
}

func (pm *VipBundleProcessManager) OnFlightBooked(ctx context.Context, event *events.FlightBooked_v1) error {
	vipBundleID := entities.MustParseVipBundleID(event.ReferenceID)

	vipBundle, err := pm.repository.Get(ctx, vipBundleID)
	if err != nil {
		return fmt.Errorf("failed to get vip bundle %s: %w", vipBundleID, err)
	}

	isInboundFlight := event.FlightID == vipBundle.InboundFlightID
	isReturnFlight := event.FlightID == vipBundle.ReturnFlightID

	if isInboundFlight {
		vipBundle, err = pm.repository.UpdateByID(
			ctx,
			vipBundleID,
			func(vipBundle entities.VipBundle) (entities.VipBundle, error) {
				vipBundle.InboundFlightTicketsIDs = event.TicketIDs
				return vipBundle, nil
			},
		)
		if err != nil {
			return fmt.Errorf("failed to update vip bundle %s: %w", vipBundleID, err)
		}

		idempotencyKey := uuid.NewString()
		cmd := commands.NewBookFlight(
			vipBundle.CustomerEmail,
			vipBundle.ReturnFlightID,
			vipBundle.Passengers,
			vipBundle.VipBundleID.String(),
			idempotencyKey,
		)

		err = pm.commandBus.Send(ctx, cmd)
		if err != nil {
			return fmt.Errorf("failed to send BookFlight command for return flight: %w", err)
		}
	} else if isReturnFlight {
		vipBundle, err = pm.repository.UpdateByID(
			ctx,
			vipBundleID,
			func(vipBundle entities.VipBundle) (entities.VipBundle, error) {
				vipBundle.ReturnFlightTicketsIDs = event.TicketIDs
				now := time.Now()
				vipBundle.ReturnFlightBookedAt = &now
				return vipBundle, nil
			},
		)
		if err != nil {
			return fmt.Errorf("failed to update vip bundle %s: %w", vipBundleID, err)
		}

		idempotencyKey := uuid.NewString()
		cmd := commands.NewBookTaxi(
			vipBundle.CustomerEmail,
			vipBundle.Passengers[0],
			vipBundle.NumberOfTickets,
			vipBundle.VipBundleID.String(),
			idempotencyKey,
		)

		err = pm.commandBus.Send(ctx, cmd)
		if err != nil {
			return fmt.Errorf("failed to send BookTaxi command: %w", err)
		}
	} else {
		return fmt.Errorf("unknown flight id %s for vip bundle %s", event.FlightID, vipBundleID)
	}

	return nil
}

func (pm *VipBundleProcessManager) OnFlightBookingFailed(ctx context.Context, event *events.FlightBookingFailed_v1) error {
	vipBundleID := entities.MustParseVipBundleID(event.ReferenceID)

	vipBundle, err := pm.repository.Get(ctx, vipBundleID)
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

	// Cancel inbound flight if return flight booking failed
	if len(vipBundle.InboundFlightTicketsIDs) > 0 {
		cmd := commands.NewCancelFlightTickets(vipBundle.InboundFlightTicketsIDs)

		err = pm.commandBus.Send(ctx, cmd)
		if err != nil {
			return fmt.Errorf("failed to send CancelFlightTickets command: %w", err)
		}
	}

	// Refund show tickets
	for _, ticketID := range vipBundle.TicketIDs {
		cmd := commands.NewRefundTicket(ticketID)
		err = pm.commandBus.Send(ctx, cmd)
		if err != nil {
			return fmt.Errorf("failed to send RefundTicket command for ticket %s: %w", ticketID, err)
		}
	}

	vipBundle, err = pm.repository.UpdateByID(
		ctx,
		vipBundleID,
		func(vipBundle entities.VipBundle) (entities.VipBundle, error) {
			vipBundle.IsFinalized = true
			vipBundle.Failed = true
			return vipBundle, nil
		},
	)
	if err != nil {
		return fmt.Errorf("failed to update vip bundle %s: %w", vipBundleID, err)
	}

	finalizedEvent := events.NewVipBundleFinalized(vipBundle.VipBundleID, false)
	err = pm.eventBus.Publish(ctx, finalizedEvent)
	if err != nil {
		return fmt.Errorf("failed to publish VipBundleFinalized_v1 event: %w", err)
	}

	return nil
}

func (pm *VipBundleProcessManager) OnTaxiBooked(ctx context.Context, event *events.TaxiBooked_v1) error {
	vipBundleID := entities.MustParseVipBundleID(event.ReferenceID)

	vipBundle, err := pm.repository.UpdateByID(
		ctx,
		vipBundleID,
		func(vipBundle entities.VipBundle) (entities.VipBundle, error) {
			taxiBookingID := event.TaxiBookingID
			vipBundle.TaxiBookingID = &taxiBookingID
			now := time.Now()
			vipBundle.TaxiBookedAt = &now
			vipBundle.IsFinalized = true
			return vipBundle, nil
		},
	)
	if err != nil {
		return fmt.Errorf("failed to update vip bundle %s: %w", vipBundleID, err)
	}

	finalizedEvent := events.NewVipBundleFinalized(vipBundle.VipBundleID, true)
	err = pm.eventBus.Publish(ctx, finalizedEvent)
	if err != nil {
		return fmt.Errorf("failed to publish VipBundleFinalized_v1 event: %w", err)
	}

	return nil
}

func (pm *VipBundleProcessManager) OnTaxiBookingFailed(ctx context.Context, event *events.TaxiBookingFailed_v1) error {
	vipBundleID := entities.MustParseVipBundleID(event.ReferenceID)

	vipBundle, err := pm.repository.Get(ctx, vipBundleID)
	if err != nil {
		return fmt.Errorf("failed to get vip bundle %s: %w", vipBundleID, err)
	}

	// Refund show tickets
	for _, ticketID := range vipBundle.TicketIDs {
		cmd := commands.NewRefundTicket(ticketID)

		err = pm.commandBus.Send(ctx, cmd)
		if err != nil {
			return fmt.Errorf("failed to send RefundTicket command for ticket %s: %w", ticketID, err)
		}
	}

	// Cancel inbound flight tickets
	if len(vipBundle.InboundFlightTicketsIDs) > 0 {
		cmd := commands.NewCancelFlightTickets(vipBundle.InboundFlightTicketsIDs)

		err = pm.commandBus.Send(ctx, cmd)
		if err != nil {
			return fmt.Errorf("failed to send CancelFlightTickets command for inbound flight: %w", err)
		}
	}

	// Cancel return flight tickets
	if len(vipBundle.ReturnFlightTicketsIDs) > 0 {
		cmd := commands.NewCancelFlightTickets(vipBundle.ReturnFlightTicketsIDs)

		err = pm.commandBus.Send(ctx, cmd)
		if err != nil {
			return fmt.Errorf("failed to send CancelFlightTickets command for return flight: %w", err)
		}
	}

	vipBundle, err = pm.repository.UpdateByID(
		ctx,
		vipBundleID,
		func(vipBundle entities.VipBundle) (entities.VipBundle, error) {
			vipBundle.IsFinalized = true
			vipBundle.Failed = true
			return vipBundle, nil
		},
	)
	if err != nil {
		return fmt.Errorf("failed to update vip bundle %s: %w", vipBundleID, err)
	}

	finalizedEvent := events.NewVipBundleFinalized(vipBundle.VipBundleID, false)
	err = pm.eventBus.Publish(ctx, finalizedEvent)
	if err != nil {
		return fmt.Errorf("failed to publish VipBundleFinalized_v1 event: %w", err)
	}

	return nil
}
