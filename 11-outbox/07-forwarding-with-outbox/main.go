package main

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/forwarder"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	watermillSQL "github.com/ThreeDotsLabs/watermill-sql/v3/pkg/sql"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"context"
)

func RunForwarder(
	db *sqlx.DB,
	rdb *redis.Client,
	outboxTopic string,
	logger watermill.LoggerAdapter,
) error {
	postgresSub, err := watermillSQL.NewSubscriber(
		db,
		watermillSQL.SubscriberConfig{
			SchemaAdapter:  watermillSQL.DefaultPostgreSQLSchema{},
			OffsetsAdapter: watermillSQL.DefaultPostgreSQLOffsetsAdapter{},
		},
		logger,
	)
	if err != nil {
		return err
	}

	err = postgresSub.SubscribeInitialize(outboxTopic)
	if err != nil {
		return err
	}

	redisPub, err := redisstream.NewPublisher(redisstream.PublisherConfig{
		Client: rdb,
	}, logger)
	if err != nil {
		return err
	}

	// TODO: your code goes here
	fwd, err := forwarder.NewForwarder(
		postgresSub,
    	redisPub,
    	logger,
    	forwarder.Config{
    	    ForwarderTopic: outboxTopic,
    	},
	)
	go func() {
    err := fwd.Run(context.Background())
    if err != nil {
        panic(err)
    }
	}()

	<-fwd.Running()




	return nil
}
