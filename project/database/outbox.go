package database

import (
	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/forwarder"
	"github.com/ThreeDotsLabs/watermill/message"
	watermillSQL "github.com/ThreeDotsLabs/watermill-sql/v3/pkg/sql"
	"github.com/jmoiron/sqlx"
	"tickets/tracing"
)

func NewOutboxPublisher(tx *sqlx.Tx, logger watermill.LoggerAdapter) (message.Publisher, error) {
	var publisher message.Publisher
	sqlPublisher, err := watermillSQL.NewPublisher(
		tx,
		watermillSQL.PublisherConfig{
			SchemaAdapter: watermillSQL.DefaultPostgreSQLSchema{},
		},
		logger,
	)
	if err != nil {
		return nil, err
	}

	publisher = log.CorrelationPublisherDecorator{Publisher: sqlPublisher}
	publisher = tracing.NewTracePublisher(publisher)

	publisher = forwarder.NewPublisher(publisher, forwarder.PublisherConfig{
		ForwarderTopic: OutboxTopic,
	})

	publisher = log.CorrelationPublisherDecorator{Publisher: publisher}
	publisher = tracing.NewTracePublisher(publisher)

	return publisher, nil
}
