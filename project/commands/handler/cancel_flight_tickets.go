package handler

import (
	"context"
	"fmt"

	"tickets/entities/commands"
)

func (h *CommandHandler) CancelFlightTickets(ctx context.Context, cmd *commands.CancelFlightTickets) error {
	err := h.transportationService.CancelFlightTickets(ctx, cmd.FlightTicketIDs)
	if err != nil {
		return fmt.Errorf("failed to cancel flight tickets: %w", err)
	}

	return nil
}
