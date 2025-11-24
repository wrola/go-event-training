package event

import (
	"context"
	"log/slog"
	"tickets/entities"
)

func (h MessageHandler) AppendToTracker(ctx context.Context, payload entities.TicketBookingConfirmed) error {
	slog.Info("Appending to tracker", "ticket_id", payload.TicketID, "customer_email", payload.CustomerEmail)

	err := h.spreadsheetsAPI.AppendRow(ctx, "tickets-to-print", []string{payload.TicketID, payload.CustomerEmail, payload.Price.Amount, payload.Price.Currency})
	if err != nil {
		return err
	}

	return nil
}