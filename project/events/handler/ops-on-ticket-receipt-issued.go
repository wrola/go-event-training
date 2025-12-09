package handler

import (
	"context"
	"fmt"
	"tickets/entities/events"
	"tickets/entities/models"
)

func (h *MessageHandler) OnTicketReceiptIssued_v1(ctx context.Context, event *events.TicketReceiptIssued_v1) error {
	return h.opsBookingRepo.UpdateReadModelByTicketID(ctx, event.TicketID, func(opsBooking *models.OpsBooking) error {
		ticket, exists := opsBooking.Tickets[event.TicketID]
		if !exists {
			return fmt.Errorf("ticket %s not found in booking %s (will retry)", event.TicketID, opsBooking.BookingID)
		}

		ticket.ReceiptNumber = event.ReceiptNumber
		ticket.ReceiptIssuedAt = event.IssuedAt

		opsBooking.Tickets[event.TicketID] = ticket
		return nil
	})
}
