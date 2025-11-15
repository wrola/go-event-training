package main

import (
	"os"
	"fmt"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/redis/go-redis/v9"
	"context"
)

func main() {
	logger := watermill.NewSlogLogger(nil)

	rdb := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
	})

	sub, err := redisstream.NewSubscriber(redisstream.SubscriberConfig{
		Client: rdb,
	}, logger)

	if err != nil {
		panic(err)
	}

	router := message.NewDefaultRouter(logger)

	if err != nil {
		panic(err)
	}

	router.AddConsumerHandler(
		"temperature_converter",
		"temperature-fahrenheit",
		sub,
		func(msg *message.Message) error {
			temperature := string(msg.Payload)
			fmt.Printf("Temperature read: %s\n", temperature)
			return nil
		},
	)

	if err := router.Run(context.Background()); err != nil {
		panic(err)
	}
}
