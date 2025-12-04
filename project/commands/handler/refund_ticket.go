package handler

import (
	"context"
	"log/slog"

	"tickets/entities/commands"
)

func (h CommandHandler) RefundTicket(ctx context.Context, command *commands.RefundTicket) error {
	slog.Info("Processing ticket refund", "ticket_id", command.TicketID)

	err := h.receiptsService.VoidReceipt(
		ctx,
		command.TicketID,
		"ticket refund",
		command.Header.IdempotencyKey,
	)
	if err != nil {
		return err
	}

	slog.Info("Ticket refund processed successfully", "ticket_id", command.TicketID)

	return nil
}
