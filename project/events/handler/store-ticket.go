package handler

import (
	"context"
	"tickets/entities/events"
	"tickets/entities/models"
	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
)

func (h MessageHandler) StoreTicket(ctx context.Context, event *events.TicketBookingConfirmed) error {
	log.FromContext(ctx).Info("Storing ticket")

	return h.ticketRepository.Add(ctx, models.Ticket{
		TicketID:      event.TicketID,
		Price:         event.Price,
		CustomerEmail: event.CustomerEmail,
	})
}