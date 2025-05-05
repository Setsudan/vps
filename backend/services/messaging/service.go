// services/messaging/service.go
package messaging

import (
	"context"
	"encoding/json"
	"time"

	m "launay-dot-one/models"
	"launay-dot-one/repositories"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

type service struct {
	redisClient *redis.Client
	repo        *repositories.MessagingRepository
}

// NewService wires up Redis + GORM for messaging.
func NewService(
	redisClient *redis.Client,
	repo *repositories.MessagingRepository,
) Service {
	return &service{redisClient: redisClient, repo: repo}
}

// SendMessage marshals the message into Redis with a 3-minute TTL.
func (s *service) SendMessage(ctx context.Context, msg *m.Message) error {
	msg.ID = uuid.NewString()
	msg.CreatedAt = time.Now()

	raw, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return s.redisClient.Set(ctx, "message:"+msg.ID, raw, 3*time.Minute).Err()
}

// AddReaction pulls the JSON message from Redis, updates its Reactions,
// then writes it back (resetting TTL to 3m).
func (s *service) AddReaction(ctx context.Context, messageID, reaction, userID string) error {
	key := "message:" + messageID
	raw, err := s.redisClient.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}

	var msg m.Message
	if err := json.Unmarshal(raw, &msg); err != nil {
		return err
	}

	// existing reactions → map[string][]string
	var rm map[string][]string
	if len(msg.Reactions) > 0 {
		if err := json.Unmarshal(msg.Reactions, &rm); err != nil {
			return err
		}
	} else {
		rm = make(map[string][]string)
	}

	// skip if already present
	for _, uid := range rm[reaction] {
		if uid == userID {
			return nil
		}
	}
	rm[reaction] = append(rm[reaction], userID)

	newReacts, err := json.Marshal(rm)
	if err != nil {
		return err
	}
	msg.Reactions = newReacts

	updated, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return s.redisClient.Set(ctx, key, updated, 3*time.Minute).Err()
}

// GetChannelHistory retrieves all messages persisted for a channel.
func (s *service) GetChannelHistory(ctx context.Context, channelID string) ([]m.Message, error) {
	return s.repo.GetMessages(ctx, channelID, "channel")
}

// TransferExpiredMessages scans Redis-keys, persists any >3m old,
// then deletes them from Redis.
func (s *service) TransferExpiredMessages(ctx context.Context) error {
	keys, err := s.redisClient.Keys(ctx, "message:*").Result()
	if err != nil {
		return err
	}

	for _, key := range keys {
		raw, err := s.redisClient.Get(ctx, key).Bytes()
		if err != nil {
			continue
		}

		var msg m.Message
		if err := json.Unmarshal(raw, &msg); err != nil {
			continue
		}
		if time.Since(msg.CreatedAt) < 3*time.Minute {
			continue
		}

		// Persist and delete
		_ = s.repo.SaveMessage(ctx, msg)
		_ = s.redisClient.Del(ctx, key).Err()
	}

	return nil
}

// GetMessagesBetweenUsers proxies to your MessagingRepository.
func (s *service) GetMessagesBetweenUsers(
	ctx context.Context, userA, userB string,
) ([]m.Message, error) {
	return s.repo.GetMessagesBetweenUsers(ctx, userA, userB)
}

// GetUserConversations returns every message where user is sender or target.
func (s *service) GetUserConversations(
	ctx context.Context, userID string,
) ([]m.Message, error) {
	return s.repo.GetAllMessagesForUser(ctx, userID)
}
