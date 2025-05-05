package listeners

import (
	"context"
	"log"
	"strings"

	"launay-dot-one/services/messaging"

	"github.com/go-redis/redis/v8"
)

func RedisExpiredListener(ctx context.Context, rdb *redis.Client, msgSvc messaging.Service) error {
	// Subscribe to expired events on DB 0.
	pubsub := rdb.PSubscribe(ctx, "__keyevent@0__:expired")
	if _, err := pubsub.Receive(ctx); err != nil {
		return err
	}
	ch := pubsub.Channel()

	go func() {
		for {
			select {
			case evt, ok := <-ch:
				if !ok {
					log.Println("Redis expired listener channel closed")
					return
				}
				if !strings.HasPrefix(evt.Payload, "message:") {
					continue
				}
				log.Printf("Expired key detected: %s", evt.Payload)
				if err := msgSvc.TransferExpiredMessages(ctx); err != nil {
					log.Printf("Error transferring expired messages: %v", err)
				} else {
					log.Printf("Successfully transferred expired messages for key: %s", evt.Payload)
				}
			case <-ctx.Done():
				log.Println("Shutting down Redis expired listener")
				return
			}
		}
	}()
	return nil
}
