package main

import( 
	"os"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/redis/go-redis/v9"
)



func main() {

	logger := watermill.NewStdLogger(false, false)

	redisClient := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
	})

	publisher, err := redisstream.NewPublisher(
		redisstream.PublisherConfig{
			Client: redisClient,
		}, logger)

	if err != nil {
		panic(err)
	}
	fiftyMessage := message.NewMessage(
		watermill.NewUUID(),
		[]byte("50"),
	)

	houndredMessage := message.NewMessage(
		watermill.NewUUID(),
		[]byte("100"),
	)

	errPublish := publisher.Publish("progress", fiftyMessage, houndredMessage)
	
	if errPublish != nil {
		panic(errPublish)
	}
}
