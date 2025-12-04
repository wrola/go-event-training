package handler

import (
	"context"
	"fmt"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
	"tickets/entities/events"
)

func (h MessageHandler) StoreTicketFile(ctx context.Context, event *events.TicketBookingConfirmed) error {
	fileID := fmt.Sprintf("%s-ticket.html", event.TicketID)

	fileContent := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>Ticket %s</title>
</head>
<body>
    <h1>Ticket Confirmation</h1>
    <p><strong>Ticket ID:</strong> %s</p>
    <p><strong>Customer Email:</strong> %s</p>
    <p><strong>Price:</strong> %s %s</p>
</body>
</html>`, event.TicketID, event.TicketID, event.CustomerEmail, event.Price.Amount, event.Price.Currency)

	err := h.filesAPI.StoreFile(ctx, fileID, fileContent)
	if err != nil {
		return fmt.Errorf("failed to store ticket file: %w", err)
	}

	log.FromContext(ctx).With("file_id", fileID).Info("Ticket file stored successfully")

	ticketPrintedEvent := events.NewTicketPrinted(event.TicketID, fileID)

	errBus := h.eventBus.Publish(ctx, ticketPrintedEvent)
	if errBus != nil {
		return fmt.Errorf("failed to publish ticket printed event: %w", errBus)
	}

	return nil
}

