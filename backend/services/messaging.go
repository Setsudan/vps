package services

import (
	"context"
	"encoding/json"
	"log"
	"sort"
	"time"

	"launay-dot-one/models"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

// DBClient abstracts PostgreSQL operations for messages.
type DBClient interface {
	GetMessages(ctx context.Context, targetID, targetType string) ([]models.Message, error)
	SaveMessage(ctx context.Context, msg models.Message) error

	// GetAllMessagesForUser retrieves all messages for a user.
	GetMessagesBetweenUsers(ctx context.Context, userA, userB string) ([]models.Message, error)
	GetAllMessagesForUser(ctx context.Context, userID string) ([]models.Message, error)
}

// MessagingService defines methods for messaging.
type MessagingService interface {
	// SendMessage stores the message in Redis temporarily.
	SendMessage(ctx context.Context, msg models.Message) error
	// AddReaction updates the reactions for a message in Redis.
	AddReaction(ctx context.Context, messageID, reaction, userID string) error
	// GetChatHistory retrieves persisted messages from PostgreSQL.
	GetChatHistory(ctx context.Context, targetID, targetType string) ([]models.Message, error)
	// TransferExpiredMessages moves messages from Redis to PostgreSQL.
	TransferExpiredMessages(ctx context.Context) error
	// GetUserConversations retrieves all conversations for a user.
	// GetMessagesBetweenUsers retrieves messages between two users.
	GetMessagesBetweenUsers(ctx context.Context, userA, userB string) ([]models.Message, error)
	// GetAllUserConversations retrieves all messages for a user.
	GetUserConversations(ctx context.Context, userID string) ([]models.Message, error)
}

type messagingService struct {
	redisClient *redis.Client
	db          DBClient
}

// NewMessagingService creates a new instance.
func NewMessagingService(redisClient *redis.Client, db DBClient) MessagingService {
	return &messagingService{
		redisClient: redisClient,
		db:          db,
	}
}

// SendMessage stores a new message in Redis with a TTL of 3 minutes.
func (ms *messagingService) SendMessage(ctx context.Context, msg models.Message) error {
	// Set unique ID and timestamp
	msg.ID = uuid.New().String()
	msg.Timestamp = time.Now()

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	key := "message:" + msg.ID
	return ms.redisClient.Set(ctx, key, data, 0).Err()
}

// AddReaction updates the Reactions field for a message in Redis.
// It unmarshals the existing JSON into a map, updates it, then re-marshals.
func (ms *messagingService) AddReaction(ctx context.Context, messageID, reaction, userID string) error {
	key := "message:" + messageID
	data, err := ms.redisClient.Get(ctx, key).Result()
	if err != nil {
		return err
	}

	var msg models.Message
	if err := json.Unmarshal([]byte(data), &msg); err != nil {
		return err
	}

	// Unmarshal the Reactions JSON into a map.
	var reactionsMap map[string][]string
	if len(msg.Reactions) > 0 {
		if err := json.Unmarshal(msg.Reactions, &reactionsMap); err != nil {
			return err
		}
	} else {
		reactionsMap = make(map[string][]string)
	}

	// Check if the user already reacted.
	for _, uid := range reactionsMap[reaction] {
		if uid == userID {
			return nil // already reacted
		}
	}
	// Append the user to the reaction slice.
	reactionsMap[reaction] = append(reactionsMap[reaction], userID)

	// Re-marshal the updated reactions map to JSON.
	newReactions, err := json.Marshal(reactionsMap)
	if err != nil {
		return err
	}
	msg.Reactions = newReactions

	updated, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return ms.redisClient.Set(ctx, key, updated, 1*time.Minute).Err()
}

// GetChatHistory retrieves messages for a given target from PostgreSQL.
func (ms *messagingService) GetChatHistory(ctx context.Context, userID, targetID string) ([]models.Message, error) {
	messages, err := ms.db.GetMessagesBetweenUsers(ctx, userID, targetID)
	if err != nil {
		return nil, err
	}

	sort.Slice(messages, func(i, j int) bool {
		return messages[i].Timestamp.Before(messages[j].Timestamp)
	})

	return messages, nil
}

// TransferExpiredMessages iterates over stored messages in Redis,
// saves them to PostgreSQL, and removes them from Redis.
func (ms *messagingService) TransferExpiredMessages(ctx context.Context) error {
	keys, err := ms.redisClient.Keys(ctx, "message:*").Result()
	if err != nil {
		return err
	}

	for _, key := range keys {
		data, err := ms.redisClient.Get(ctx, key).Result()
		if err != nil {
			continue
		}

		var msg models.Message
		if err := json.Unmarshal([]byte(data), &msg); err != nil {
			continue
		}

		// Check if the message is older than 3 minutes.
		if time.Since(msg.Timestamp) < 3*time.Minute {
			continue
		}

		if err := ms.db.SaveMessage(ctx, msg); err != nil {
			continue
		}

		if err := ms.redisClient.Del(ctx, key).Err(); err != nil {
			continue
		}
		log.Printf("Transferred expired message: %s", key)
	}
	return nil
}

func (ms *messagingService) GetMessagesBetweenUsers(ctx context.Context, userA, userB string) ([]models.Message, error) {
	return ms.db.GetMessagesBetweenUsers(ctx, userA, userB)
}

// Duplicate method removed.
func (ms *messagingService) GetUserConversations(ctx context.Context, userID string) ([]models.Message, error) {
	// Retrieve all messages for the user using the DB client.
	allMsgs, err := ms.db.GetAllMessagesForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	return allMsgs, nil
}
