package handler

import (
	"context"
	"fmt"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/clients/dead_nation"
	"github.com/google/uuid"
	"tickets/entities/events"
)

func (h MessageHandler) BookingMadeUpdate(ctx context.Context, payload *events.BookingMade_v1) error {
	show, err := h.showsRepository.ShowByID(ctx, payload.ShowID)
	if err != nil {
		return fmt.Errorf("failed to get show by ID %s: %w", payload.ShowID, err)
	}

	bookingUUID, err := uuid.Parse(payload.BookingID)
	if err != nil {
		return fmt.Errorf("failed to parse booking ID as UUID: %w", err)
	}

	eventUUID, err := uuid.Parse(show.DeadNationID)
	if err != nil {
		return fmt.Errorf("failed to parse event ID as UUID: %w", err)
	}

	resp, err := h.deadNationAPI.PostTicketBookingWithResponse(
		ctx,
		dead_nation.PostTicketBookingJSONRequestBody{
			BookingId:       bookingUUID,
			EventId:         eventUUID, 
			NumberOfTickets: payload.NumberOfTickets,
			CustomerAddress: payload.CustomerEmail,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to call Dead Nation API: %w", err)
	}

	if resp.StatusCode() != 200 && resp.StatusCode() != 201 {
		return fmt.Errorf("Dead Nation API returned status %d", resp.StatusCode())
	}

	return nil
}
