package event

import (
	"context"
	"log/slog"
	"tickets/entities"
)

func (h MessageHandler) AppendToTracker(ctx context.Context, payload entities.TicketBookingConfirmed) error {
	slog.Info("Appending to tracker", "ticket_id", payload.TicketID, "customer_email", payload.CustomerEmail)

	// Fix missing currency
	if payload.Price.Currency == "" {
		slog.Warn("Added missing currency", "ticket_id", payload.TicketID)
		payload.Price = entities.Money{
			Amount:   payload.Price.Amount,
			Currency: "USD",
		}
	}

	err := h.spreadsheetsAPI.AppendRow(ctx, "tickets-to-print", []string{payload.TicketID, payload.CustomerEmail, payload.Price.Amount, payload.Price.Currency})
	if err != nil {
		return err
	}

	return nil
}