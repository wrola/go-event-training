package handler

import (
	"context"
	"time"
	"tickets/entities/events"
	"tickets/entities/models"
	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
)

func (h MessageHandler) StoreTicket(ctx context.Context, event *events.TicketBookingConfirmed_v1) error {
	log.FromContext(ctx).Info("Storing ticket")

	now := time.Now()
	return h.ticketRepository.Add(ctx, models.Ticket{
		TicketID:      event.TicketID,
		Price:         event.Price,
		CustomerEmail: event.CustomerEmail,
		ConfirmedAt:   now.Format(time.RFC3339),
		RefundedAt:    "", // Explicitly set to empty string
	})
}