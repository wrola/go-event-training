package event

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
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
	Remove(ctx context.Context, ticket entities.Ticket) error
}

type FilesAPI interface {
	StoreFile(ctx context.Context, fileID string, fileContent string) error
}

type MessageHandler struct {
	spreadsheetsAPI  SpreadsheetsAPI
	receiptsService  ReceiptsService
	ticketRepository TicketRepository
	filesAPI         FilesAPI
	eventBus 		*cqrs.EventBus
}

func NewMessageHandler(spreadsheetsAPI SpreadsheetsAPI, receiptsService ReceiptsService, ticketRepository TicketRepository, filesAPI FilesAPI, eventBus *cqrs.EventBus) *MessageHandler {
	if spreadsheetsAPI == nil {
		panic("missing spreadsheetsAPI")
	}
	if receiptsService == nil {
		panic("missing receiptsService")
	}
	if ticketRepository == nil {
		panic("Missing ticket repository")
	}
	if filesAPI == nil {
		panic("missing filesAPI")
	}
	if eventBus == nil {
		panic("Missing eventBus")
	}

	return &MessageHandler{
		spreadsheetsAPI:  spreadsheetsAPI,
		receiptsService:  receiptsService,
		ticketRepository: ticketRepository,
		filesAPI:         filesAPI,
		eventBus:         eventBus,
	}
}


