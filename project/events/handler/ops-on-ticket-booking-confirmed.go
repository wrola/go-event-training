package handler

import (
	"context"
	"fmt"
	"time"
	"tickets/entities/events"
	"tickets/entities/models"
	"github.com/google/uuid"
)

func (h *MessageHandler) OnTicketBookingConfirmed_v1(ctx context.Context, event *events.TicketBookingConfirmed_v1) error {
	bookingID, err := uuid.Parse(event.BookingID)
	if err != nil {
		return fmt.Errorf("invalid booking_id: %w", err)
	}

	return h.opsBookingRepo.UpdateReadModelByBookingID(ctx, bookingID.String(), func(opsBooking *models.OpsBooking) error {
		if opsBooking.Tickets == nil {
			opsBooking.Tickets = make(map[string]models.OpsTicket)
		}

		ticket := opsBooking.Tickets[event.TicketID]

		ticket.PriceAmount = event.Price.Amount
		ticket.PriceCurrency = event.Price.Currency
		ticket.CustomerEmail = event.CustomerEmail

		now := time.Now()
		ticket.ConfirmedAt = &now

		opsBooking.Tickets[event.TicketID] = ticket
		return nil
	})
}
