package database

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/forwarder"
	"github.com/ThreeDotsLabs/watermill/message"
	watermillSQL "github.com/ThreeDotsLabs/watermill-sql/v3/pkg/sql"
	"github.com/jmoiron/sqlx"
)

const OutboxTopic = "events_to_forward"

func NewForwarder(
	db *sqlx.DB,
	publisher message.Publisher,
	logger watermill.LoggerAdapter,
) (*forwarder.Forwarder, error) {
	sqlSubscriber, err := watermillSQL.NewSubscriber(
		db,
		watermillSQL.SubscriberConfig{
			SchemaAdapter:    watermillSQL.DefaultPostgreSQLSchema{},
			OffsetsAdapter:   watermillSQL.DefaultPostgreSQLOffsetsAdapter{},
			InitializeSchema: true, 
		},
		logger,
	)
	if err != nil {
		return nil, err
	}

	err = sqlSubscriber.SubscribeInitialize(OutboxTopic)
	if err != nil {
		return nil, err
	}

	fwd, err := forwarder.NewForwarder(
		sqlSubscriber,
		publisher,
		logger,
		forwarder.Config{
			ForwarderTopic: OutboxTopic,
		},
	)
	if err != nil {
		return nil, err
	}

	return fwd, nil
}
