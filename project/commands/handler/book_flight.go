package handler

import (
	"context"
	"fmt"

	"tickets/adapters"
	"tickets/entities/commands"
	"tickets/entities/events"
)

func (h *CommandHandler) BookFlight(ctx context.Context, cmd *commands.BookFlight) error {
	ticketIDs, err := h.transportationService.BookFlight(
		ctx,
		cmd.CustomerEmail,
		cmd.FlightID,
		cmd.Passengers,
		cmd.ReferenceID,
		cmd.IdempotencyKey,
	)

	if err != nil {
		if permanentErr, ok := err.(*adapters.PermanentFlightBookingError); ok {
			failedEvent := events.NewFlightBookingFailed(
				cmd.FlightID,
				permanentErr.Error(),
				cmd.ReferenceID,
			)

			if publishErr := h.eventBus.Publish(ctx, failedEvent); publishErr != nil {
				return fmt.Errorf("failed to publish FlightBookingFailed event: %w", publishErr)
			}

			return nil
		}

		return fmt.Errorf("failed to book flight: %w", err)
	}

	bookedEvent := events.NewFlightBooked(cmd.FlightID, ticketIDs, cmd.ReferenceID)
	if err := h.eventBus.Publish(ctx, bookedEvent); err != nil {
		return fmt.Errorf("failed to publish FlightBooked event: %w", err)
	}

	return nil
}
