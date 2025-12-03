package main

import (
	watermillSQL "github.com/ThreeDotsLabs/watermill-sql/v3/pkg/sql"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func PublishInTx(
	message *message.Message,
	tx *sqlx.Tx,
) error {
	
	publisher, err := watermillSQL.NewPublisher(
		tx, 
		watermillSQL.PublisherConfig{
			SchemaAdapter: watermillSQL.DefaultPostgreSQLSchema{},
		},
		watermill.NewSlogLogger(nil),
	)
	if err != nil {
		panic(err)
	}

	publisher.Publish("ItemAddedToCart", message)
	return nil
}
