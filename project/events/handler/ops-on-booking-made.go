package handler

import (
	"context"
	"fmt"
	"time"
	"tickets/entities/events"
	"github.com/google/uuid"
)

func (h *MessageHandler) OnBookingMade(ctx context.Context, event *events.BookingMade_v1) error {
	bookingID, err := uuid.Parse(event.BookingID)
	if err != nil {
		return fmt.Errorf("invalid booking_id: %w", err)
	}

	return h.opsBookingRepo.CreateBooking(ctx, bookingID.String(), time.Now())
}
