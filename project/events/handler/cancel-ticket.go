package handler

import (
	"context"
	"log/slog"
	"tickets/entities/events"
)

func (h MessageHandler) CancelTicket(ctx context.Context, payload *events.TicketBookingCanceled_v1) error {
	slog.Info("Cancelling ticket", "ticket_id", payload.TicketID, "customer_email", payload.CustomerEmail)

	err := h.spreadsheetsAPI.AppendRow(ctx, "tickets-to-refund", []string{payload.TicketID, payload.CustomerEmail, payload.Price.Amount, payload.Price.Currency})
	if err != nil {
		return err
	}
	err = h.ticketRepository.SoftDelete(ctx, payload.TicketID)
	if err != nil {
		return err
	}
	return nil
}	