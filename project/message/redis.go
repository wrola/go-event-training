package message

import (
	"os"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/redis/go-redis/v9"
	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
)
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

	decoratedPublisher := log.CorrelationPublisherDecorator{Publisher: publisher}

	return decoratedPublisher, nil
}

func NewEventBus(pub message.Publisher) (*cqrs.EventBus, error) {

	bus, err := cqrs.NewEventBusWithConfig(
		pub,
		cqrs.EventBusConfig{
			GeneratePublishTopic: func(params cqrs.GenerateEventPublishTopicParams) (string, error) {
				return params.EventName, nil
			},
			Marshaler: cqrs.JSONMarshaler{
					GenerateName: cqrs.StructName,
			},
		},
	)

	if err != nil {
		return nil, err
	}


	return bus, nil
}