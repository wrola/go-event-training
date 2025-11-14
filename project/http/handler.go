package http

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/message"
)

type SpreadsheetsAPI interface {
	AppendRow(ctx context.Context, sheetName string, row []string) error
}

type ReceiptsService interface {
	IssueReceipt(ctx context.Context, ticketID string) error
}

type Handler struct {
	publisher message.Publisher
}
