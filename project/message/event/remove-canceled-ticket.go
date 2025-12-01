package event 

import ( 
	"context"
	"tickets/entities"
	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
)

func (h MessageHandler) RemoveCanceledTicket (ctx context.Context, event *entities.TicketBookingCanceled ) error {
	log.FromContext(ctx).Info("Removal canceled Ticket from Store")

	return h.ticketRepository.Remove(ctx, entities.Ticket{
		TicketID:      event.TicketID,
	})
}
