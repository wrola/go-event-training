package handler

import (
	"context"
	"tickets/entities/events"
	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
)

func (h MessageHandler) RemoveCanceledTicket (ctx context.Context, event *events.TicketBookingCanceled_v1 ) error {
	log.FromContext(ctx).Info("Soft deleting canceled ticket from store")

	return h.ticketRepository.SoftDelete(ctx, event.TicketID)
}
