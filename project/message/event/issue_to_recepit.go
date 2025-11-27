package event

import (
	"context"
	"log/slog"
	"tickets/entities"
)

func (h MessageHandler) IssueReceipt(ctx context.Context, payload *entities.TicketBookingConfirmed) error {
	slog.Info("Issuing receipt", "ticket_id", payload.TicketID, "customer_email", payload.CustomerEmail)

	err := h.receiptsService.IssueReceipt(ctx, *payload)
	if err != nil {
		return err
	}

	return nil
}