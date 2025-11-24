package event

import (
	"context"
	"tickets/entities"
)

type SpreadsheetsAPI interface {
	AppendRow(ctx context.Context, sheetName string, row []string) error
}

type ReceiptsService interface {
	IssueReceipt(ctx context.Context, payload entities.TicketBookingConfirmed) error
}

type MessageHandler struct {
	spreadsheetsAPI SpreadsheetsAPI
	receiptsService ReceiptsService
}

func NewMessageHandler(spreadsheetsAPI SpreadsheetsAPI, receiptsService ReceiptsService) *MessageHandler {
	if spreadsheetsAPI == nil {
		panic("missing spreadsheetsAPI")
	}
	if receiptsService == nil {
		panic("missing receiptsService")
	}

	return &MessageHandler{
		spreadsheetsAPI: spreadsheetsAPI,
		receiptsService: receiptsService,
	}
}


