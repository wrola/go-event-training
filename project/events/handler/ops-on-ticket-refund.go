package handler

import (
	"context"
	"fmt"
	"time"
	"tickets/entities/events"
	"tickets/entities/models"
)

func (h *MessageHandler) OnTicketRefunded_v1(ctx context.Context, event *events.TicketRefunded_v1) error {
	return h.opsBookingRepo.UpdateReadModelByTicketID(ctx, event.TicketID, func(opsBooking *models.OpsBooking) error {
		ticket, exists := opsBooking.Tickets[event.TicketID]
		if !exists {
			return fmt.Errorf("ticket %s not found in booking %s (will retry)", event.TicketID, opsBooking.BookingID)
		}

		now := time.Now()
		ticket.RefundedAt = &now

		opsBooking.Tickets[event.TicketID] = ticket
		return nil
	})
}
