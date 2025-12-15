package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/IBM/sarama"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/urfave/cli/v2"
)

const PoisonQueueTopic = "PoisonQueue"

type Message struct {
	ID     string
	Reason string
}

type Handler struct {
	subscriber message.Subscriber
	publisher  message.Publisher
}

func NewHandler() (*Handler, error) {
	logger := watermill.NewSlogLogger(nil)

	cfg := sarama.NewConfig()
	cfg.Consumer.Offsets.Initial = sarama.OffsetOldest

	sub, err := kafka.NewSubscriber(
		kafka.SubscriberConfig{
			Brokers:               []string{os.Getenv("KAFKA_ADDR")},
			Unmarshaler:           kafka.DefaultMarshaler{},
			ConsumerGroup:         "poison-queue-cli",
			OverwriteSaramaConfig: cfg,
		},
		logger,
	)
	if err != nil {
		return nil, err
	}

	pub, err := kafka.NewPublisher(
		kafka.PublisherConfig{
			Brokers:   []string{os.Getenv("KAFKA_ADDR")},
			Marshaler: kafka.DefaultMarshaler{},
		},
		logger,
	)
	if err != nil {
		return nil, err
	}

	return &Handler{
		subscriber: sub,
		publisher:  pub,
	}, nil
}

func (h *Handler) Preview(ctx context.Context) ([]Message, error) {
	var messages []Message

	err := h.iterate(ctx, func(msg *message.Message) (bool, error) {
		messages = append(messages, Message{
			ID:     msg.UUID,
			Reason: msg.Metadata.Get(middleware.ReasonForPoisonedKey),
		})
		return true, nil
	})
	if err != nil {
		return nil, err
	}

	return messages, nil
}

func (h *Handler) Remove(ctx context.Context, messageID string) error {
	found := false

	err := h.iterate(ctx, func(msg *message.Message) (bool, error) {
		if msg.UUID == messageID {
			found = true
			return false, nil 
		}
		return true, nil 
	})
	if err != nil {
		return err
	}

	if !found {
		return fmt.Errorf("message with ID %s not found", messageID)
	}

	return nil
}

func (h *Handler) Requeue(ctx context.Context, messageID string) error {
	found := false

	err := h.iterate(ctx, func(msg *message.Message) (bool, error) {
		if msg.UUID == messageID {
			found = true

			originalTopic := msg.Metadata.Get(middleware.PoisonedTopicKey)

			if originalTopic == "" {
				return true, fmt.Errorf("message %s has no original topic in metadata", messageID)
			}

			err := h.publisher.Publish(originalTopic, msg)
			if err != nil {
				return true, fmt.Errorf("failed to publish message %s to topic %s: %w", messageID, originalTopic, err)
			}

			return false, nil
		}

		return true, nil
	})

	if err != nil {
		return err
	}

	if !found {
		return fmt.Errorf("message with ID %s not found", messageID)
	}

	return nil
}

func (h *Handler) iterate(ctx context.Context, actionFunc func(msg *message.Message) (bool, error)) error {
	logger := watermill.NewSlogLogger(nil)
	router := message.NewDefaultRouter(logger)

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	firstMessageUUID := ""

	done := false

	router.AddHandler(
		"preview",
		PoisonQueueTopic,
		h.subscriber,
		PoisonQueueTopic,
		h.publisher,
		func(msg *message.Message) ([]*message.Message, error) {
			if done {
				cancel()
				return nil, errors.New("done")
			}

			if firstMessageUUID == "" {
				firstMessageUUID = msg.UUID
			} else if firstMessageUUID == msg.UUID {
				// we've read all messages
				done = true
				return nil, errors.New("done")
			}

			keep, err := actionFunc(msg)
			if err != nil {
				return nil, err
			}

			if !keep {
				if msg.UUID == firstMessageUUID {
					done = true
				}
				return nil, nil
			}

			return []*message.Message{msg}, nil
		},
	)

	err := router.Run(ctx)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	app := &cli.App{
		Name:  "poison-queue-cli",
		Usage: "Manage the Poison Queue",
		Commands: []*cli.Command{
			{
				Name:  "preview",
				Usage: "preview messages",
				Action: func(c *cli.Context) error {
					h, err := NewHandler()
					if err != nil {
						return err
					}

					messages, err := h.Preview(c.Context)
					if err != nil {
						return err
					}

					for _, m := range messages {
						fmt.Printf("%v\t%v\n", m.ID, m.Reason)
					}

					return nil
				},
			},
			{
				Name:      "remove",
				ArgsUsage: "<message_id>",
				Usage:     "remove message",
				Action: func(c *cli.Context) error {
					h, err := NewHandler()
					if err != nil {
						return err
					}

					err = h.Remove(c.Context, c.Args().First())
					if err != nil {
						return err
					}

					return nil
				},
			},
			{
				Name:      "requeue",
				ArgsUsage: "<message_id>",
				Usage:     "requeue message",
				Action: func(c *cli.Context) error {
					h, err := NewHandler()
					if err != nil {
						return err
					}

					err = h.Requeue(c.Context, c.Args().First())
					if err != nil {
						return err
					}

					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
