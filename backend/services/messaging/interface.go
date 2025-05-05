package messaging

import (
	"context"

	m "launay-dot-one/models"
)

// Service defines all messaging operations.
type Service interface {
	// SendMessage enqueues a new message (in Redis) before persistence.
	SendMessage(ctx context.Context, msg *m.Message) error

	// AddReaction adds a reaction to an in-flight message in Redis.
	AddReaction(ctx context.Context, messageID, reaction, userID string) error

	// GetChannelHistory loads all persisted messages for a channel.
	GetChannelHistory(ctx context.Context, channelID string) ([]m.Message, error)

	// TransferExpiredMessages moves aged messages from Redis → PostgreSQL.
	TransferExpiredMessages(ctx context.Context) error

	// GetMessagesBetweenUsers loads the one-on-one history between two users.
	GetMessagesBetweenUsers(ctx context.Context, userA, userB string) ([]m.Message, error)

	// GetUserConversations loads every message where the user is sender or recipient.
	GetUserConversations(ctx context.Context, userID string) ([]m.Message, error)
}
