package message

import (
	"os"
	"github.com/ThreeDotsLabs/watermill"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/redis/go-redis/v9"
	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
)

func NewRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
	})
}

func NewLogger() watermill.LoggerAdapter {
	return watermill.NewStdLogger(false, false)
}

func NewMessagePublisher(redisClient *redis.Client, logger watermill.LoggerAdapter) (message.Publisher, error) {

	publisher, err := redisstream.NewPublisher(
		redisstream.PublisherConfig{
			Client: redisClient,
		}, logger)

	if err != nil {
		return nil, err
	}

	decoratedPublisher := log.CorrelationPublisherDecorator{Publisher: publisher}

	return decoratedPublisher, nil
}

func NewSubscriberConstructor(redisClient *redis.Client, logger watermill.LoggerAdapter) cqrs.EventProcessorSubscriberConstructorFn {
	return func(params cqrs.EventProcessorSubscriberConstructorParams) (message.Subscriber, error) {
		
		consumerGroup := "receipt-workers"
		if params.HandlerName == "AppendToTracker" || params.HandlerName == "CancelTicket" {
			consumerGroup = "tracker-workers"
		}

		return redisstream.NewSubscriber(
			redisstream.SubscriberConfig{
				Client:        redisClient,
				ConsumerGroup: consumerGroup,
			}, logger)
	}
}
