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
}

func NewMessageHandler(spreadsheetsAPI SpreadsheetsAPI, receiptsService ReceiptsService) *MessageHandler {
	return &MessageHandler{
		spreadsheetsAPI: spreadsheetsAPI,
		receiptsService: receiptsService,
	}
}

func NewMessageConsumers() (message.Subscriber, message.Subscriber, error) {
	logger := watermill.NewStdLogger(false, false)

	redisClient := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
	})

	subRecepit, err := redisstream.NewSubscriber(
		redisstream.SubscriberConfig{
			Client:        redisClient,
			ConsumerGroup: "receipt-workers",
		}, logger)
	if err != nil {
		return nil, nil, err
	}

	subTracker, err := redisstream.NewSubscriber(
		redisstream.SubscriberConfig{
			Client:        redisClient,
			ConsumerGroup: "tracker-workers",
		}, logger)
	if err != nil {
		return nil, nil, err
	}

	return subRecepit, subTracker, nil
}

// RunSubscribers starts goroutines to process messages from the subscribers
func (h *MessageHandler) RunSubscribers(ctx context.Context, subRecepit, subTracker message.Subscriber) error {
	go h.processMessages(ctx, subRecepit, subTracker)

	return nil
}

// processMessages subscribes to a topic and processes incoming messages
func (h *MessageHandler) processMessages(ctx context.Context, subTracker message.Subscriber, subRecepit message.Subscriber) {
	logger := watermill.NewSlogLogger(nil)

	router := message.NewDefaultRouter(logger)

	router.AddConsumerHandler(
		"handler_ticket_receipt",
		"issue-receipt",
		subRecepit,
		func(msg *message.Message) error {
				ticketID := string(msg.Payload)
				err := h.receiptsService.IssueReceipt(ctx, ticketID)
				if err != nil {
					return err
				}
			
				return nil
			},
	)
	router.AddConsumerHandler(
		"handler_append_to_tracker",
		"append-to-tracker",
		subTracker,
		func(msg *message.Message) error {
				ticketID := string(msg.Payload)
				err := h.spreadsheetsAPI.AppendRow(ctx, "tickets-to-print", []string{ticketID})
				if err != nil {
					return err
				}
			
				return nil
			},
	)	


	if err := router.Run(ctx); err != nil {
		panic(err)
	}	
	
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

