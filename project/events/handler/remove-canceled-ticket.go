package handler

import ( 
	"context"
	"tickets/entities/events"
	"tickets/entities/models"
	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
)

func (h MessageHandler) RemoveCanceledTicket (ctx context.Context, event *events.TicketBookingCanceled ) error {
	log.FromContext(ctx).Info("Removal canceled Ticket from Store")

	return h.ticketRepository.Remove(ctx, models.Ticket{
		TicketID:      event.TicketID,
	})
}
