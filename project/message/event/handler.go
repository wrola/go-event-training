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

type TicketRepository interface {
	Add(ctx context.Context, ticket entities.Ticket) error
}

type MessageHandler struct {
	spreadsheetsAPI SpreadsheetsAPI
	receiptsService ReceiptsService
	ticketRepository TicketRepository
}

func NewMessageHandler(spreadsheetsAPI SpreadsheetsAPI, receiptsService ReceiptsService, ticketRepository TicketRepository) *MessageHandler {
	if spreadsheetsAPI == nil {
		panic("missing spreadsheetsAPI")
	}
	if receiptsService == nil {
		panic("missing receiptsService")
	}
	if ticketRepository == nil { 
		panic("Missing ticket repository")
	}

	return &MessageHandler{
		spreadsheetsAPI: spreadsheetsAPI,
		receiptsService: receiptsService,
		ticketRepository: ticketRepository,
	}
}


