package realtime

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

// PresenceService defines methods for updating user presence.
type PresenceService interface {
	SetStatus(ctx context.Context, userID string, status string) error
	// Optionally: GetStatus(ctx context.Context, userID string) (string, error)
}

type presenceService struct {
	redisClient *redis.Client
}

// NewPresenceService creates a new PresenceService.
func NewPresenceService(redisClient *redis.Client) PresenceService {
	return &presenceService{redisClient: redisClient}
}

// SetStatus stores the user's presence in Redis with a TTL.
func (ps *presenceService) SetStatus(ctx context.Context, userID string, status string) error {
	// You can use a TTL to auto-expire stale status entries.
	return ps.redisClient.Set(ctx, "presence:"+userID, status, 5*time.Minute).Err()
}