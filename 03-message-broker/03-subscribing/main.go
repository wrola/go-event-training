package main
import ( 
	"fmt"
	"os"
	"context"
	"time"
	"encoding/json"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	// "github.com/ThreeDotsLabs/watermill/message"
	"github.com/redis/go-redis/v9"
)

func main() {
	parent := context.Background()
  	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	logger := watermill.NewStdLogger(false, false)
	
	redisClient := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
	})

	subcriber, err := redisstream.NewSubscriber(
		redisstream.SubscriberConfig{
			Client: redisClient,
		}, logger)

	if err != nil {
		panic(err)
	}


	messages, err := subcriber.Subscribe(ctx, "progress")

	if err != nil {
		panic(err)
	}

	for msg := range messages {
		num := 0
		err := json.Unmarshal(msg.Payload, &num)
  			
		if err != nil {
			fmt.Printf("cannot unmarshal message: %v", err)
			msg.Nack()
			continue
		}	
		fmt.Printf("Message ID: %s - %v \n", msg.UUID, num)

		msg.Ack()
	}


	defer cancel()
}
