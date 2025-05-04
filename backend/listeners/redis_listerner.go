package listeners

import (
	"context"
	"log"
	"strings"

	"launay-dot-one/services"

	"github.com/go-redis/redis/v8"
)

// RedisExpiredListener subscribes to Redis key expiration events and triggers
// the transfer of expired message keys to PostgreSQL.
// Make sure that Redis is configured with keyspace notifications enabled,
// e.g., set notify-keyspace-events to at least "Ex" in the Redis configuration.
func RedisExpiredListener(ctx context.Context, rdb *redis.Client, messagingService services.MessagingService) error {
	// Subscribe to expired events on DB 0.
	pubsub := rdb.PSubscribe(ctx, "__keyevent@0__:expired")
	if _, err := pubsub.Receive(ctx); err != nil {
		return err
	}

	// Obtain the channel for receiving messages.
	ch := pubsub.Channel()

	// Start processing the key expiration messages in a separate goroutine.
	go func() {
		for {
			select {
			case msg, ok := <-ch:
				if !ok {
					log.Println("Redis expired listener channel closed")
					return
				}
				// Process only keys that belong to our message namespace.
				if strings.HasPrefix(msg.Payload, "message:") {
					log.Printf("Redis expired event detected for key: %s", msg.Payload)
					// Call our messaging service method to transfer expired messages.
					// Note: In this design, TransferExpiredMessages loops over remaining keys,
					// so even if one key expires, it will check all keys.
					if err := messagingService.TransferExpiredMessages(ctx); err != nil {
						log.Printf("Error transferring expired message(s): %v", err)
					} else {
						log.Printf("Expired message(s) transferred successfully for key: %s", msg.Payload)
					}
				}
			case <-ctx.Done():
				log.Println("Shutting down Redis expired listener")
				return
			}
		}
	}()

	return nil
}
