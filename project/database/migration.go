package database

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"tickets/entities/events"
	"tickets/entities/models"
)

type ReadModelMigrator struct {
	eventRepo EventRepository
	opsRepo   OpsBookingReadModelRepository
	logger    *log.Logger
}

func NewReadModelMigrator(
	eventRepo EventRepository,
	opsRepo OpsBookingReadModelRepository,
) *ReadModelMigrator {
	return &ReadModelMigrator{
		eventRepo: eventRepo,
		opsRepo:   opsRepo,
		logger:    log.New(os.Stdout, "[MIGRATION] ", log.LstdFlags),
	}
}

func (m *ReadModelMigrator) MigrateReadModel(ctx context.Context) error {
	m.logger.Println("Starting read model migration...")

	if err := m.waitForEvents(ctx); err != nil {
		return err
	}

	storedEvents, err := m.eventRepo.GetAllEvents(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch events: %w", err)
	}

	m.logger.Printf("Found %d events to process", len(storedEvents))

	for i, storedEvent := range storedEvents {
		if err := m.processEvent(ctx, storedEvent); err != nil {
			return fmt.Errorf("failed to process event %d (%s): %w",
				i, storedEvent.EventName, err)
		}

		if (i+1)%100 == 0 {
			m.logger.Printf("Processed %d/%d events", i+1, len(storedEvents))
		}
	}

	m.logger.Printf("Migration completed successfully. Processed %d events", len(storedEvents))
	return nil
}

func (m *ReadModelMigrator) waitForEvents(ctx context.Context) error {
	m.logger.Println("Waiting for events table to be populated...")

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	timeout := time.After(5 * time.Minute)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			return fmt.Errorf("timeout waiting for events table to be populated")
		case <-ticker.C:
			count, err := m.eventRepo.CountEvents(ctx)
			if err != nil {
				m.logger.Printf("Error counting events: %v", err)
				continue
			}

			if count > 0 {
				m.logger.Printf("Events table populated with %d events", count)
				return nil
			}

			m.logger.Printf("Events table empty, waiting...")
		}
	}
}

func (m *ReadModelMigrator) processEvent(ctx context.Context, storedEvent StoredEvent) error {
	switch storedEvent.EventName {
	case "BookingMade_v0":
		return m.handleBookingMade_v0(ctx, storedEvent.EventPayload)
	case "TicketBookingConfirmed_v0":
		return m.handleTicketBookingConfirmed_v0(ctx, storedEvent.EventPayload)
	case "TicketReceiptIssued_v0":
		return m.handleTicketReceiptIssued_v0(ctx, storedEvent.EventPayload)
	case "TicketPrinted_v0":
		return m.handleTicketPrinted_v0(ctx, storedEvent.EventPayload)
	case "TicketRefunded_v0":
		return m.handleTicketRefunded_v0(ctx, storedEvent.EventPayload)
	default:
		m.logger.Printf("Skipping unknown event type: %s", storedEvent.EventName)
		return nil
	}
}

func (m *ReadModelMigrator) handleBookingMade_v0(ctx context.Context, payload json.RawMessage) error {
	var v0Event events.BookingMade_v0
	if err := json.Unmarshal(payload, &v0Event); err != nil {
		return fmt.Errorf("failed to unmarshal BookingMade_v0: %w", err)
	}

	bookingID := v0Event.BookingID.String()
	bookedAt := v0Event.Header.PublishedAt

	return m.opsRepo.CreateBooking(ctx, bookingID, bookedAt)
}

func (m *ReadModelMigrator) handleTicketBookingConfirmed_v0(ctx context.Context, payload json.RawMessage) error {
	var v0Event events.TicketBookingConfirmed_v0
	if err := json.Unmarshal(payload, &v0Event); err != nil {
		return fmt.Errorf("failed to unmarshal TicketBookingConfirmed_v0: %w", err)
	}

	return m.opsRepo.UpdateReadModelByBookingID(ctx, v0Event.BookingID, func(opsBooking *models.OpsBooking) error {
		if opsBooking.Tickets == nil {
			opsBooking.Tickets = make(map[string]models.OpsTicket)
		}

		ticket := opsBooking.Tickets[v0Event.TicketID]
		ticket.PriceAmount = v0Event.Price.Amount
		ticket.PriceCurrency = v0Event.Price.Currency
		ticket.CustomerEmail = v0Event.CustomerEmail

		confirmedAt := v0Event.Header.PublishedAt
		ticket.ConfirmedAt = &confirmedAt

		opsBooking.Tickets[v0Event.TicketID] = ticket
		return nil
	})
}

func (m *ReadModelMigrator) handleTicketReceiptIssued_v0(ctx context.Context, payload json.RawMessage) error {
	var v0Event events.TicketReceiptIssued_v0
	if err := json.Unmarshal(payload, &v0Event); err != nil {
		return fmt.Errorf("failed to unmarshal TicketReceiptIssued_v0: %w", err)
	}

	return m.opsRepo.UpdateReadModelByTicketID(ctx, v0Event.TicketID, func(opsBooking *models.OpsBooking) error {
		ticket, exists := opsBooking.Tickets[v0Event.TicketID]
		if !exists {
			return fmt.Errorf("ticket %s not found in booking %s", v0Event.TicketID, opsBooking.BookingID)
		}

		ticket.ReceiptNumber = v0Event.ReceiptNumber
		ticket.ReceiptIssuedAt = v0Event.IssuedAt

		opsBooking.Tickets[v0Event.TicketID] = ticket
		return nil
	})
}

func (m *ReadModelMigrator) handleTicketPrinted_v0(ctx context.Context, payload json.RawMessage) error {
	var v0Event events.TicketPrinted_v0
	if err := json.Unmarshal(payload, &v0Event); err != nil {
		return fmt.Errorf("failed to unmarshal TicketPrinted_v0: %w", err)
	}

	return m.opsRepo.UpdateReadModelByTicketID(ctx, v0Event.TicketID, func(opsBooking *models.OpsBooking) error {
		ticket, exists := opsBooking.Tickets[v0Event.TicketID]
		if !exists {
			return fmt.Errorf("ticket %s not found in booking %s", v0Event.TicketID, opsBooking.BookingID)
		}

		ticket.PrintedFileName = v0Event.FileName
		ticket.PrintedAt = v0Event.Header.PublishedAt

		opsBooking.Tickets[v0Event.TicketID] = ticket
		return nil
	})
}

func (m *ReadModelMigrator) handleTicketRefunded_v0(ctx context.Context, payload json.RawMessage) error {
	var v0Event events.TicketRefunded_v0
	if err := json.Unmarshal(payload, &v0Event); err != nil {
		return fmt.Errorf("failed to unmarshal TicketRefunded_v0: %w", err)
	}

	return m.opsRepo.UpdateReadModelByTicketID(ctx, v0Event.TicketID, func(opsBooking *models.OpsBooking) error {
		ticket, exists := opsBooking.Tickets[v0Event.TicketID]
		if !exists {
			return fmt.Errorf("ticket %s not found in booking %s", v0Event.TicketID, opsBooking.BookingID)
		}

		refundedAt := v0Event.Header.PublishedAt
		ticket.RefundedAt = &refundedAt

		opsBooking.Tickets[v0Event.TicketID] = ticket
		return nil
	})
}
