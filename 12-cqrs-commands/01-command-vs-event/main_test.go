// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package main

import (
	"context"
	"testing"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type CapturedMessage struct {
	NotificationID string
	Email          string
	Message        string
}

type StubSender struct {
	Messages []CapturedMessage
}

func (m *StubSender) SendNotification(ctx context.Context, notificationID, email, message string) error {
	m.Messages = append(m.Messages, CapturedMessage{
		NotificationID: notificationID,
		Email:          email,
		Message:        message,
	})
	return nil
}

func Test(t *testing.T) {
	watermillLogger := watermill.NewSlogLogger(nil)

	pubSub := gochannel.NewGoChannel(gochannel.Config{}, watermillLogger)

	commandBus, err := cqrs.NewCommandBusWithConfig(
		pubSub,
		cqrs.CommandBusConfig{
			GeneratePublishTopic: func(params cqrs.CommandBusGeneratePublishTopicParams) (string, error) {
				return "commands", nil
			},
			Marshaler: cqrs.JSONMarshaler{
				GenerateName: cqrs.StructName,
			},
			Logger: watermillLogger,
		},
	)
	if err != nil {
		panic(err)
	}

	notificationsStub := &StubSender{}

	router := message.NewDefaultRouter(watermillLogger)

	processor := NewProcessor(router, notificationsStub, pubSub, watermillLogger)
	assert.IsType(t, &cqrs.CommandProcessor{}, processor)

	go func() {
		if err := router.Run(context.Background()); err != nil {
			panic(err)
		}
	}()
	<-router.Running()

	cmd := &SendNotification{
		NotificationID: uuid.NewString(),
		Email:          "email@example.com",
		Message:        "Welcome!",
	}
	err = commandBus.Send(context.Background(), cmd)
	require.NoError(t, err)

	assert.EventuallyWithT(
		t,
		func(t *assert.CollectT) {
			assert.Equal(
				t,
				[]CapturedMessage{
					{
						NotificationID: cmd.NotificationID,
						Email:          cmd.Email,
						Message:        cmd.Message,
					},
				},
				notificationsStub.Messages,
			)
		},
		1*time.Second,
		100*time.Millisecond,
	)
}
