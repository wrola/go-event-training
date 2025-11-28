package event

import (
	"context"
	"tickets/entities"
	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
)

func (h MessageHandler) StoreTicket(ctx context.Context, event *entities.TicketBookingConfirmed) error {
	log.FromContext(ctx).Info("Storing ticket")

	return h.ticketRepository.Add(ctx, entities.Ticket{
		TicketID:      event.TicketID,
		Price:         event.Price,
		CustomerEmail: event.CustomerEmail,
	})
}