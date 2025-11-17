package message

import (
	"context"
	"os"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/redis/go-redis/v9"
)

type MessagePublisher interface {
	Publish(topic string, msg *message.Message) error
}

type MessageSubscriber interface {
	Subscribe(ctx context.Context, topic string) (<-chan *message.Message, error)
}

type SpreadsheetsAPI interface {
	AppendRow(ctx context.Context, sheetName string, row []string) error
}

type ReceiptsService interface {
	IssueReceipt(ctx context.Context, ticketID string) error
}

type Task int

const (
	TaskIssueReceipt Task = iota
	TaskAppendToTracker
)

// MessageHandler processes messages from subscribers using injected dependencies
type MessageHandler struct {
	spreadsheetsAPI SpreadsheetsAPI
	receiptsService ReceiptsService
	router		  *message.Router
	subRecepit	  message.Subscriber
	subTracker	  message.Subscriber
}

func NewMessageHandler(spreadsheetsAPI SpreadsheetsAPI, receiptsService ReceiptsService) *MessageHandler {
	return &MessageHandler{
		spreadsheetsAPI: spreadsheetsAPI,
		receiptsService: receiptsService,
	}
}

func(h *MessageHandler) NewMessageConsumers() (error) {
	logger := watermill.NewStdLogger(false, false)

	redisClient := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
	})

	var err error
	h.subRecepit, err = redisstream.NewSubscriber(
		redisstream.SubscriberConfig{
			Client:        redisClient,
			ConsumerGroup: "receipt-workers",
		}, logger)
	if err != nil {
		return err
	}

	h.subTracker, err = redisstream.NewSubscriber(
		redisstream.SubscriberConfig{
			Client:        redisClient,
			ConsumerGroup: "tracker-workers",
		}, logger)
	if err != nil {
		return err
	}

	return nil
}


func (h *MessageHandler) StartProcessMessages(ctx context.Context) error {
	return h.router.Run(ctx)
}

func (h *MessageHandler) AddConsumers(ctx context.Context) error{
	logger := watermill.NewSlogLogger(nil)

	h.router = message.NewDefaultRouter(logger)

	h.router.AddConsumerHandler(
		"handler_ticket_receipt",
		"issue-receipt",
		h.subRecepit,
		func(msg *message.Message) error {
				ticketID := string(msg.Payload)
				err := h.receiptsService.IssueReceipt(ctx, ticketID)
				if err != nil {
					return err
				}
			
				return nil
			},
	)
	h.router.AddConsumerHandler(
		"handler_append_to_tracker",
		"append-to-tracker",
		h.subTracker,
		func(msg *message.Message) error {
				ticketID := string(msg.Payload)
				err := h.spreadsheetsAPI.AppendRow(ctx, "tickets-to-print", []string{ticketID})
				if err != nil {
					return err
				}
			
				return nil
			},
	)	
	return nil
}

func NewMessageProducer() (message.Publisher, error) {
	logger := watermill.NewStdLogger(false, false)

	redisClient := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
	})

	publisher, err := redisstream.NewPublisher(
		redisstream.PublisherConfig{
			Client: redisClient,
		}, logger)
	if err != nil {
		return nil, err
	}

	return publisher, nil
}

func (h *MessageHandler) Running() <-chan struct{} {
	return h.router.Running()
}