package adapters

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"tickets/entities/events"
	"github.com/ThreeDotsLabs/go-event-driven/v2/common/clients"
	"github.com/ThreeDotsLabs/go-event-driven/v2/common/clients/receipts"
)

type ReceiptsServiceClient struct {
	// we are not mocking this client: it's pointless to use interface here
	clients *clients.Clients
}

func NewReceiptsServiceClient(clients *clients.Clients) *ReceiptsServiceClient {
	if clients == nil {
		panic("NewReceiptsServiceClient: clients is nil")
	}

	return &ReceiptsServiceClient{clients: clients}
}

func (c ReceiptsServiceClient) IssueReceipt(ctx context.Context, request events.TicketBookingConfirmed) error {
	idempotencyKey := request.Header.IdempotencyKey
	resp, err := c.clients.Receipts.PutReceiptsWithResponse(ctx, receipts.CreateReceipt{
		IdempotencyKey: &idempotencyKey,
		TicketId:       request.TicketID,
		Price: receipts.Money{
			MoneyAmount:   request.Price.Amount,
			MoneyCurrency: request.Price.Currency,
		},
	})

	if err != nil {
		return fmt.Errorf("failed to post receipt: %w", err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		// receipt already exists
		return nil
	case http.StatusCreated:
		// receipt was created
		return nil
	default:
		return fmt.Errorf("unexpected status code for POST receipts-api/receipts: %d", resp.StatusCode())
	}
}

func (c ReceiptsServiceClient) VoidReceipt(ctx context.Context, ticketID string, reason string, idempotencyKey string) error {
	resp, err := c.clients.Receipts.PutVoidReceiptWithResponse(ctx, receipts.VoidReceiptRequest{
		TicketId:     ticketID,
		Reason:        "customer requested refund",
		IdempotentId: &idempotencyKey,
	})

	if err != nil {
		return fmt.Errorf("failed to void receipt: %w", err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return nil
	case http.StatusNoContent:
		return nil
	default:
		return fmt.Errorf("unexpected status code for PUT void-receipt: %d", resp.StatusCode())
	}
}

type ReceiptsServiceClientStub struct {
	IssuedReceipts []events.TicketBookingConfirmed
	lock sync.Mutex
}

func (r *ReceiptsServiceClientStub) IssueReceipt(ctx context.Context, request events.TicketBookingConfirmed) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.IssuedReceipts =  append(r.IssuedReceipts, request)

	return nil
}

func (r *ReceiptsServiceClientStub) VoidReceipt(ctx context.Context, ticketID string, reason string, idempotencyKey string) error {
	return nil
}