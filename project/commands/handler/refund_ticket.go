package handler

import (
	"context"
	"log/slog"

	"tickets/entities/commands"
	"tickets/entities/events"
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

	err = h.paymentsService.RefundPayment(
		ctx,
		commands.PaymentRefund{
			TicketID:       command.TicketID,
			RefundReason:   "customer requested refund",
			IdempotencyKey: command.Header.IdempotencyKey,
		},
	)
	if err != nil {
		return err
	}

	slog.Info("Ticket refund processed successfully", "ticket_id", command.TicketID)
	event := events.NewTicketRefund(command.TicketID)

	err = h.eventBus.Publish(ctx, event)

	if err != nil {
		return err 
	}
	
	return nil
}
