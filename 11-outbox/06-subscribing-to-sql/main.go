package main

import (
	watermillSQL "github.com/ThreeDotsLabs/watermill-sql/v3/pkg/sql"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"context"
)

func SubscribeToMessages(
	db *sqlx.DB,
	topic string,
	logger watermill.LoggerAdapter,
) (<-chan *message.Message, error) {
	subscriber, err := watermillSQL.NewSubscriber(
		db,
		watermillSQL.SubscriberConfig{
			SchemaAdapter:  watermillSQL.DefaultPostgreSQLSchema{},
			OffsetsAdapter: watermillSQL.DefaultPostgreSQLOffsetsAdapter{},
		},
		logger,
	)

	if err != nil {
		panic(err)
	}

	err = subscriber.SubscribeInitialize(topic)
	if err != nil {
		panic(err)
	}

	var subErr error
	messgaes, subErr := subscriber.Subscribe(context.Background(), topic)
	if subErr != nil {
		panic(subErr)
	}

	return messgaes, nil
}
