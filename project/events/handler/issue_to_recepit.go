package handler

import (
	"context"
	"log/slog"
	"tickets/entities/events"
)

func (h MessageHandler) IssueReceipt(ctx context.Context, payload *events.TicketBookingConfirmed) error {
	slog.Info("Issuing receipt", "ticket_id", payload.TicketID, "customer_email", payload.CustomerEmail)

	resp, err := h.receiptsService.IssueReceipt(ctx, *payload)
	if err != nil {
		return err
	}

	receiptIssuedEvent := events.NewTicketReceiptIssued(
		payload.TicketID,
		resp.ReceiptNumber,
		resp.IssuedAt,
		payload.Header.IdempotencyKey,
	)

	err = h.eventBus.Publish(ctx, receiptIssuedEvent)
	if err != nil {
		return err
	}

	slog.Info("Receipt issued successfully", "ticket_id", payload.TicketID, "receipt_number", resp.ReceiptNumber)
	return nil
}