package event

import (
	"context"
	"log/slog"
	"tickets/entities"
)

func (h MessageHandler) IssueReceipt(ctx context.Context, payload entities.TicketBookingConfirmed) error {
	slog.Info("Issuing receipt", "ticket_id", payload.TicketID, "customer_email", payload.CustomerEmail)

	if payload.Price.Currency == "" {
		slog.Warn("Added missing currency", "ticket_id", payload.TicketID)
		payload.Price = entities.Money{
			Amount:   payload.Price.Amount,
			Currency: "USD",
		}
	}

	err := h.receiptsService.IssueReceipt(ctx, payload)
	if err != nil {
		return err
	}

	return nil
}