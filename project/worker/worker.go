package worker

import (
	"context"
	"log/slog"
)

type Task int

const (
	TaskIssueReceipt Task = iota
	TaskAppendToTracker
)

type Message struct {
	Task     Task
	TicketID string
}

type Worker struct {
	queue chan Message

	spreadsheetsAPI SpreadsheetsAPI
	receiptsService ReceiptsService
}

type SpreadsheetsAPI interface {
	AppendRow(ctx context.Context, sheetName string, row []string) error
}

type ReceiptsService interface {
	IssueReceipt(ctx context.Context, ticketID string) error
}

func NewWorker(
	spreadsheetsAPI SpreadsheetsAPI,
	receiptsService ReceiptsService,
) *Worker {
	return &Worker{
		queue: make(chan Message, 100),

		spreadsheetsAPI: spreadsheetsAPI,
		receiptsService: receiptsService,
	}
}

func (w *Worker) Send(msgs ...Message) {
	for _, msg := range msgs {
		w.queue <- msg
	}
}

func (w *Worker) Run(ctx context.Context) {
	for msg := range w.queue {
		switch msg.Task {
		case TaskIssueReceipt:
			err := w.receiptsService.IssueReceipt(ctx, msg.TicketID)
			if err != nil {
				w.Send(msg)
				slog.With("error", err).Error("failed to issue the receipt")
			}
		case TaskAppendToTracker:
			err := w.spreadsheetsAPI.AppendRow(ctx, "tickets-to-print", []string{msg.TicketID})
			if err != nil {
					w.Send(msg)
				slog.With("error", err).Error("failed to append to tracker")
			}
		}
	}
}
