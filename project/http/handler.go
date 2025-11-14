package http

import (
	"context"
	worker "tickets/worker"
)

type SpreadsheetsAPI interface {
	AppendRow(ctx context.Context, sheetName string, row []string) error
}

type ReceiptsService interface {
	IssueReceipt(ctx context.Context, ticketID string) error
}

type Handler struct {
	Worker *worker.Worker
}
