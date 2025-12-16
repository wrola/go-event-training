package handler

import (
	"context"
	"fmt"

	"tickets/adapters"
	"tickets/entities/commands"
	"tickets/entities/events"
)

func (h *CommandHandler) BookTaxi(ctx context.Context, cmd *commands.BookTaxi) error {
	bookingID, err := h.transportationService.BookTaxi(
		ctx,
		cmd.CustomerEmail,
		cmd.CustomerName,
		cmd.NumberOfPassengers,
		cmd.ReferenceID,
		cmd.IdempotencyKey,
	)

	if err != nil {
		if permanentErr, ok := err.(*adapters.PermanentTaxiBookingError); ok {
			failedEvent := events.NewTaxiBookingFailed(
				permanentErr.Error(),
				cmd.ReferenceID,
			)

			if publishErr := h.eventBus.Publish(ctx, failedEvent); publishErr != nil {
				return fmt.Errorf("failed to publish TaxiBookingFailed event: %w", publishErr)
			}

			return nil
		}

		return fmt.Errorf("failed to book taxi: %w", err)
	}

	bookedEvent := events.NewTaxiBooked(bookingID, cmd.ReferenceID)
	if err := h.eventBus.Publish(ctx, bookedEvent); err != nil {
		return fmt.Errorf("failed to publish TaxiBooked event: %w", err)
	}

	return nil
}
