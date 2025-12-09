package handler

import (
	"context"

	"tickets/entities"
	"tickets/entities/commands"
	"tickets/entities/events"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
)

type ReceiptsService interface {
	IssueReceipt(ctx context.Context, payload events.TicketBookingConfirmed_v1) (entities.IssueReceiptResponse, error)
	VoidReceipt(ctx context.Context, ticketID string, reason string, idempotencyKey string) error
}

type PaymentsService interface {
	RefundPayment(ctx context.Context, refundPayment commands.PaymentRefund) error
}

type CommandHandler struct {
	receiptsService ReceiptsService
	paymentsService PaymentsService
	eventBus 		*cqrs.EventBus
}

func NewCommandHandler(receiptsService ReceiptsService, paymentsService PaymentsService, eventBus *cqrs.EventBus) *CommandHandler {
	if receiptsService == nil {
		panic("receiptsService is required")
	}
	if paymentsService == nil {
		panic("paymentsService is required")
	}
	if eventBus == nil {
		panic("event bus is required")
	}

	return &CommandHandler{
		receiptsService: receiptsService,
		paymentsService: paymentsService,
		eventBus: eventBus,
	}
}
