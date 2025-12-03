package main

import (
	"database/sql"
	"fmt"

	"github.com/ThreeDotsLabs/watermill/components/forwarder"
	"github.com/ThreeDotsLabs/watermill"
	watermillSQL "github.com/ThreeDotsLabs/watermill-sql/v3/pkg/sql"
	"github.com/ThreeDotsLabs/watermill/message"
	_ "github.com/lib/pq"
)

// Don't change this variable - used in tests
const outboxTopic = "events_to_forward"

func PublishInTx(
	topic string,
	msg *message.Message,
	tx *sql.Tx,
	logger watermill.LoggerAdapter,
) error {
	var publisher message.Publisher
	var err error

	publisher, err = watermillSQL.NewPublisher(
		tx,
		watermillSQL.PublisherConfig{
			SchemaAdapter: watermillSQL.DefaultPostgreSQLSchema{},
		},
		logger,
	)
	if err != nil {
		return fmt.Errorf("failed to create outbox publisher: %w", err)
	}

	// TODO: your code goes here
	publisher = forwarder.NewPublisher(publisher, forwarder.PublisherConfig{
		ForwarderTopic: outboxTopic,
	})


	publisher.Publish(topic, msg)


	return nil
}
