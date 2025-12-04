package handler

import (
	"context"

	"tickets/entities/events"
)

type ReceiptsService interface {
	IssueReceipt(ctx context.Context, payload events.TicketBookingConfirmed) error
	VoidReceipt(ctx context.Context, ticketID string, reason string, idempotencyKey string) error
}

type CommandHandler struct {
	receiptsService ReceiptsService
}

func NewCommandHandler(receiptsService ReceiptsService) *CommandHandler {
	if receiptsService == nil {
		panic("receiptsService is required")
	}

	return &CommandHandler{
		receiptsService: receiptsService,
	}
}
