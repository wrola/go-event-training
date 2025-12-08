package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"tickets/entities/events"
	"tickets/entities/models"
)

func (h *MessageHandler) OnBookingMade(ctx context.Context, event *events.BookingMade) error {
	bookingID, err := uuid.Parse(event.BookingID)
	if err != nil {
		return fmt.Errorf("invalid booking_id: %w", err)
	}

	return h.opsBookingRepo.CreateBooking(ctx, bookingID, time.Now())
}

func (h *MessageHandler) OnTicketBookingConfirmed(ctx context.Context, event *events.TicketBookingConfirmed) error {
	bookingID, err := uuid.Parse(event.BookingID)
	if err != nil {
		return fmt.Errorf("invalid booking_id: %w", err)
	}

	return h.opsBookingRepo.UpdateReadModelByBookingID(ctx, bookingID, func(opsBooking *models.OpsBooking) error {
		if opsBooking.Tickets == nil {
			opsBooking.Tickets = make(map[string]models.OpsTicket)
		}

		ticket := opsBooking.Tickets[event.TicketID]

		ticket.PriceAmount = event.Price.Amount
		ticket.PriceCurrency = event.Price.Currency
		ticket.CustomerEmail = event.CustomerEmail
		ticket.Status = "confirmed"

		opsBooking.Tickets[event.TicketID] = ticket
		return nil
	})
}

func (h *MessageHandler) OnTicketReceiptIssued(ctx context.Context, event *events.TicketReceiptIssued) error {
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

func (h *MessageHandler) OnTicketPrinted(ctx context.Context, event *events.TicketPrinted) error {
	return h.opsBookingRepo.UpdateReadModelByTicketID(ctx, event.TicketID, func(opsBooking *models.OpsBooking) error {
		ticket, exists := opsBooking.Tickets[event.TicketID]
		if !exists {
			return fmt.Errorf("ticket %s not found in booking %s (will retry)", event.TicketID, opsBooking.BookingID)
		}

		ticket.PrintedFileName = event.FileName
		ticket.PrintedAt = time.Now()

		opsBooking.Tickets[event.TicketID] = ticket
		return nil
	})
}

func (h *MessageHandler) OnTicketRefunded(ctx context.Context, event *events.TicketRefunded) error {
	return h.opsBookingRepo.UpdateReadModelByTicketID(ctx, event.TicketID, func(opsBooking *models.OpsBooking) error {
		ticket, exists := opsBooking.Tickets[event.TicketID]
		if !exists {
			return fmt.Errorf("ticket %s not found in booking %s (will retry)", event.TicketID, opsBooking.BookingID)
		}

		ticket.Status = "refunded"

		opsBooking.Tickets[event.TicketID] = ticket
		return nil
	})
}
