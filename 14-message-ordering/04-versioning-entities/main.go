package main

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
)

type AlertTriggered struct {
	AlertID      string    `json:"alert_id"`
	AlertVersion int       `json:"alert_version"`
	TriggeredAt  time.Time `json:"triggered_at"`
}

type AlertResolved struct {
	AlertID      string    `json:"alert_id"`
	AlertVersion int       `json:"alert_version"`
	ResolvedAt   time.Time `json:"resolved_at"`
}

type AlertUpdated struct {
	AlertID     string `json:"alert_id"`
	IsTriggered bool   `json:"is_triggered"`
	LastAlertVersion int    `json:"last_alert_version"`
}

func main() {
	logger := watermill.NewSlogLogger(nil)

	kafkaAddr := os.Getenv("KAFKA_ADDR")

	router := message.NewDefaultRouter(logger)

	eventProcessor, err := cqrs.NewEventProcessorWithConfig(
		router,
		cqrs.EventProcessorConfig{
			GenerateSubscribeTopic: func(params cqrs.EventProcessorGenerateSubscribeTopicParams) (string, error) {
				return params.EventName, nil
			},
			SubscriberConstructor: func(params cqrs.EventProcessorSubscriberConstructorParams) (message.Subscriber, error) {
				sub, err := kafka.NewSubscriber(kafka.SubscriberConfig{
					Brokers:       []string{kafkaAddr},
					Unmarshaler:   kafka.DefaultMarshaler{},
					ConsumerGroup: params.HandlerName,
					// Make sure to use this config: it lets us validate your solution!
					OverwriteSaramaConfig: newConfig(),
				}, logger)
				if err != nil {
					panic(err)
				}
				return sub, nil
			},
			AckOnUnknownEvent: true,
			Marshaler:         cqrs.JSONMarshaler{},
			Logger:            logger,
		},
	)
	if err != nil {
		panic(err)
	}

	pub, err := kafka.NewPublisher(kafka.PublisherConfig{
		Brokers:   []string{kafkaAddr},
		Marshaler: kafka.DefaultMarshaler{},
	}, logger)
	if err != nil {
		panic(err)
	}

	eventBus, err := cqrs.NewEventBusWithConfig(
		pub,
		cqrs.EventBusConfig{
			GeneratePublishTopic: func(params cqrs.GenerateEventPublishTopicParams) (string, error) {
				return params.EventName, nil
			},
			Marshaler: cqrs.JSONMarshaler{},
			Logger:    logger,
		},
	)
	if err != nil {
		panic(err)
	}

	lock := sync.Mutex{}
	alerts := map[string]AlertUpdated{}

	err = eventProcessor.AddHandlers(
		cqrs.NewEventHandler("OnAlertTriggered", func(ctx context.Context, event *AlertTriggered) error {
			lock.Lock()
			defer lock.Unlock()

			alert, ok := alerts[event.AlertID]
			if !ok {
				alert = AlertUpdated{
					AlertID: event.AlertID,
					LastAlertVersion: 0,
				}
			}

			if alert.LastAlertVersion + 1 != event.AlertVersion {
				return nil
			}
			alert.IsTriggered = true
			alert.LastAlertVersion = event.AlertVersion
			alerts[event.AlertID] = alert

			return eventBus.Publish(ctx, alert)
		}),
		cqrs.NewEventHandler("OnAlertResolved", func(ctx context.Context, event *AlertResolved) error {
			lock.Lock()
			defer lock.Unlock()

			alert, ok := alerts[event.AlertID]
			if !ok {
				alert = AlertUpdated{
					AlertID: event.AlertID,
				}
			}

			if alert.LastAlertVersion + 1 != event.AlertVersion {
				return nil
			}

			alert.IsTriggered = false
			alert.LastAlertVersion = event.AlertVersion
			alerts[event.AlertID] = alert
			
			return eventBus.Publish(ctx, alert)
		}),
	)	

	if err != nil {
		panic(err)
	}

	err = router.Run(context.Background())
	if err != nil {
		panic(err)
	}
}

func newConfig() *sarama.Config {
	cfg := sarama.NewConfig()
	cfg.Consumer.Offsets.Initial = sarama.OffsetOldest
	return cfg
}
