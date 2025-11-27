package event

import (
	"context"
	"log/slog"
	"tickets/entities"
)

func (h MessageHandler) CancelTicket(ctx context.Context, payload *entities.TicketBookingCanceled) error {
	slog.Info("Cancelling ticket", "ticket_id", payload.TicketID, "customer_email", payload.CustomerEmail)

	err := h.spreadsheetsAPI.AppendRow(ctx, "tickets-to-refund", []string{payload.TicketID, payload.CustomerEmail, payload.Price.Amount, payload.Price.Currency})
	if err != nil {
		return err
	}

	return nil
}	