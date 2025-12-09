package handler

import (
	"context"
	"time"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
	"tickets/entities/events"
)

func (h MessageHandler) UpdateRefundedTicket(ctx context.Context, event *events.TicketRefunded) error {
	log.FromContext(ctx).Info("Updating refunded ticket")

	now := time.Now()
	return h.ticketRepository.UpdateRefundedAt(ctx, event.TicketID, now.Format(time.RFC3339))
}
