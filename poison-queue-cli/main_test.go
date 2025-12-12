// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package main

import (
	"context"
	"os"
	"testing"

	"github.com/IBM/sarama"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/google/uuid"
)

func Test(t *testing.T) {
	logger := watermill.NewSlogLogger(nil)

	sub, err := kafka.NewSubscriber(
		kafka.SubscriberConfig{
			Brokers:     []string{os.Getenv("KAFKA_ADDR")},
			Unmarshaler: kafka.DefaultMarshaler{},
			InitializeTopicDetails: &sarama.TopicDetail{
				NumPartitions:     1,
				ReplicationFactor: 1,
			},
		},
		logger,
	)
	if err != nil {
		t.Fatal(err)
	}

	err = sub.SubscribeInitialize(PoisonQueueTopic)
	if err != nil {
		t.Fatal(err)
	}

	pub, err := kafka.NewPublisher(
		kafka.PublisherConfig{
			Brokers:   []string{os.Getenv("KAFKA_ADDR")},
			Marshaler: kafka.DefaultMarshaler{},
		},
		logger,
	)
	if err != nil {
		t.Fatal(err)
	}

	var uuids []string
	for i := 0; i < 10; i++ {
		msg := message.NewMessage(uuid.NewString(), []byte("{}"))
		msg.Metadata.Set(middleware.ReasonForPoisonedKey, "network down")
		if err := pub.Publish(PoisonQueueTopic, msg); err != nil {
			t.Fatal(err)
		}
		uuids = append(uuids, msg.UUID)
	}

	assertMessages(t, uuids)
	// Check if no messages were lost
	assertMessages(t, uuids)
}

func assertMessages(t *testing.T, expectedUUIDs []string) {
	t.Helper()

	h, err := NewHandler()
	if err != nil {
		t.Fatal(err)
	}

	messages, err := h.Preview(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if len(messages) != len(expectedUUIDs) {
		t.Fatalf("expected %v messages, got %d", len(expectedUUIDs), len(messages))
	}

	for _, uuid := range expectedUUIDs {
		found := false
		for _, msg := range messages {
			if msg.ID == uuid {
				if msg.Reason != "network down" {
					t.Fatalf("expected reason to be 'network down', got %s", msg.Reason)
				}
				found = true
				break
			}
		}

		if !found {
			t.Fatalf("expected message with uuid %s, but not found", uuid)
		}
	}
}
