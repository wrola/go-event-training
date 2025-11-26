package main

import (
	"context"
	"sync"
	
)

type IssueReceiptRequest struct {
	TicketID string `json:"ticket_id"`
	Price    Money  `json:"price"`
}

type Money struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

type ReceiptsService interface {
	IssueReceipt(ctx context.Context, request IssueReceiptRequest) error
}

type ReceiptsServiceStub struct {
	IssuedReceipts []IssueReceiptRequest
	lock sync.Mutex
}

func (r *ReceiptsServiceStub) IssueReceipt(ctx context.Context, request IssueReceiptRequest) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.IssuedReceipts =  append(r.IssuedReceipts, request)

	return nil
}