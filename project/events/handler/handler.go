package handler

import (
	"context"
	"time"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/clients/dead_nation"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"tickets/entities"
	"tickets/entities/commands"
	"tickets/entities/events"
	"tickets/entities/models"
)

type SpreadsheetsAPI interface {
	AppendRow(ctx context.Context, sheetName string, row []string) error
}

type ReceiptsService interface {
	IssueReceipt(ctx context.Context, payload events.TicketBookingConfirmed) (entities.IssueReceiptResponse, error)
	VoidReceipt(ctx context.Context, ticketID string, reason string, idempotencyKey string) error
}

type TicketRepository interface {
	Add(ctx context.Context, ticket models.Ticket) error
	Remove(ctx context.Context, ticket models.Ticket) error
	UpdateRefundedAt(ctx context.Context, ticketID string, refundedAt string) error
	SoftDelete(ctx context.Context, ticketID string) error
}

type FilesAPI interface {
	StoreFile(ctx context.Context, fileID string, fileContent string) error
}

type ShowsRepository interface {
	ShowByID(ctx context.Context, showID string) (models.Show, error)
}

type DeadNationAPI interface {
	PostTicketBookingWithResponse(
		ctx context.Context,
		body dead_nation.PostTicketBookingJSONRequestBody,
		reqEditors ...dead_nation.RequestEditorFn,
	) (*dead_nation.PostTicketBookingResponse, error)
}

type PaymentsService interface {
	RefundPayment(ctx context.Context, refundPayment commands.PaymentRefund) error
}

type OpsBookingRepository interface {
	CreateBooking(ctx context.Context, bookingID string, bookedAt time.Time) error
	UpdateReadModelByBookingID(ctx context.Context, bookingID string, updateFn func(*models.OpsBooking) error) error
	UpdateReadModelByTicketID(ctx context.Context, ticketID string, updateFn func(*models.OpsBooking) error) error
	FindReadModelByBookingID(ctx context.Context, bookingID string) (models.OpsBooking, error)
	FindReadModelByTicketID(ctx context.Context, ticketID string) (models.OpsBooking, error)

}

type EventRepository interface {
	StoreEvent(ctx context.Context, eventID string, publishedAt time.Time, eventName string, eventPayload []byte) error
}

type MessageHandler struct {
	spreadsheetsAPI  SpreadsheetsAPI
	receiptsService  ReceiptsService
	ticketRepository TicketRepository
	filesAPI         FilesAPI
	showsRepository  ShowsRepository
	deadNationAPI    DeadNationAPI
	eventBus         *cqrs.EventBus
	opsBookingRepo   OpsBookingRepository
	eventRepo        EventRepository
}

func NewMessageHandler(
	spreadsheetsAPI SpreadsheetsAPI,
	receiptsService ReceiptsService,
	ticketRepository TicketRepository,
	filesAPI FilesAPI,
	showsRepository ShowsRepository,
	deadNationAPI DeadNationAPI,
	eventBus *cqrs.EventBus,
	opsBookingRepo OpsBookingRepository,
	eventRepo EventRepository,
) *MessageHandler {
	if spreadsheetsAPI == nil {
		panic("missing spreadsheetsAPI")
	}
	if receiptsService == nil {
		panic("missing receiptsService")
	}
	if ticketRepository == nil {
		panic("missing ticketRepository")
	}
	if filesAPI == nil {
		panic("missing filesAPI")
	}
	if showsRepository == nil {
		panic("missing showsRepository")
	}
	if deadNationAPI == nil {
		panic("missing deadNationAPI")
	}
	if eventBus == nil {
		panic("missing eventBus")
	}
	if eventRepo == nil {
		panic("missing eventRepo")
	}

	return &MessageHandler{
		spreadsheetsAPI:  spreadsheetsAPI,
		receiptsService:  receiptsService,
		ticketRepository: ticketRepository,
		filesAPI:         filesAPI,
		showsRepository:  showsRepository,
		deadNationAPI:    deadNationAPI,
		eventBus:         eventBus,
		opsBookingRepo:   opsBookingRepo,
		eventRepo:        eventRepo,
	}
}


