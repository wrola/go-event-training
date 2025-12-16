package adapters

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/clients"
	"github.com/ThreeDotsLabs/go-event-driven/v2/common/clients/transportation"
	"github.com/google/uuid"
)

type TransportationService interface {
	BookFlight(ctx context.Context, customerEmail, flightID string, passengers []string, referenceID, idempotencyKey string) ([]string, error)
}

type TransportationServiceClient struct {
	clients *clients.Clients
}

func NewTransportationServiceClient(clients *clients.Clients) *TransportationServiceClient {
	if clients == nil {
		panic("NewTransportationServiceClient: clients is nil")
	}

	return &TransportationServiceClient{clients: clients}
}

func (c TransportationServiceClient) BookFlight(
	ctx context.Context,
	customerEmail, flightID string,
	passengers []string,
	referenceID, idempotencyKey string,
) ([]string, error) {
	flightUUID, err := uuid.Parse(flightID)
	if err != nil {
		return nil, fmt.Errorf("invalid flight ID: %w", err)
	}

	resp, err := c.clients.Transportation.PutFlightTicketsWithResponse(ctx, transportation.BookFlightTicketRequest{
		CustomerEmail:  customerEmail,
		FlightId:       flightUUID,
		PassengerNames: passengers,
		ReferenceId:    referenceID,
		IdempotencyKey: idempotencyKey,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to book flight: %w", err)
	}

	switch resp.StatusCode() {
	case http.StatusCreated:
		ticketIDs := make([]string, len(resp.JSON201.TicketIds))
		for i, id := range resp.JSON201.TicketIds {
			ticketIDs[i] = id.String()
		}
		return ticketIDs, nil
	case http.StatusBadRequest:
		reason := "bad request"
		if resp.JSON400 != nil && resp.JSON400.Error != "" {
			reason = resp.JSON400.Error
		}
		return nil, &PermanentFlightBookingError{reason: reason}
	default:
		return nil, fmt.Errorf(
			"unexpected status code for PUT transportation-api/transportation/flight-tickets: %d",
			resp.StatusCode(),
		)
	}
}

type PermanentFlightBookingError struct {
	reason string
}

func (e *PermanentFlightBookingError) Error() string {
	return fmt.Sprintf("permanent flight booking error: %s", e.reason)
}

func (e *PermanentFlightBookingError) IsPermanent() bool {
	return true
}
