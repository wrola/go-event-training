// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package main

import (
	"context"
	"testing"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type StubPublisher struct {
	PublishedMessages []*message.Message
}

func (m *StubPublisher) Publish(topic string, messages ...*message.Message) error {
	m.PublishedMessages = append(m.PublishedMessages, messages...)
	return nil
}

func (m *StubPublisher) Close() error {
	return nil
}

func TestCorrelationPublisherDecorator(t *testing.T) {
	stubPublisher := &StubPublisher{}

	var publisher message.Publisher = stubPublisher
	publisher = CorrelationPublisherDecorator{publisher}

	msg := message.NewMessage(watermill.NewUUID(), nil)
	expecedCorrelationID := uuid.NewString()
	msg.SetContext(ContextWithCorrelationID(context.Background(), expecedCorrelationID))

	err := publisher.Publish("test", msg)
	require.NoError(t, err)

	require.Equal(
		t,
		1,
		len(stubPublisher.PublishedMessages),
		"one message should be published",
	)

	assert.Equal(
		t,
		expecedCorrelationID,
		stubPublisher.PublishedMessages[0].Metadata.Get("correlation_id"),
		"correlation_id should be set in metadata",
	)
}
