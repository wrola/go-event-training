package handler

import (
	"context"
	"errors"
	"fmt"

	"tickets/database"
	"tickets/entities/commands"
	"tickets/entities/models"
)

func (h *CommandHandler) BookShowTickets(ctx context.Context, cmd *commands.BookShowTickets) error {
	booking := models.NewBooking(
		cmd.BookingID,
		cmd.ShowID,
		cmd.NumberOfTickets,
		cmd.CustomerEmail,
	)

	err := h.bookingsRepository.AddBooking(ctx, booking)
	if err != nil {
		if errors.Is(err, database.ErrBookingAlreadyExists) {
			return nil
		}
		
		return fmt.Errorf("failed to add booking: %w", err)
	}

	return nil
}
