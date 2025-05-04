package services

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
)

// Location represents a user’s location and update time.
type Location struct {
	UserID    string    `json:"user_id"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Timestamp time.Time `json:"timestamp"`
}

// LocationService defines methods to update and retrieve locations.
type LocationService interface {
	// UpdateLocation writes a new location for a user with a TTL of 5 minutes.
	UpdateLocation(ctx context.Context, userID string, latitude, longitude float64) error
	// GetAllLocations retrieves all stored locations (for users that updated within the TTL).
	GetAllLocations(ctx context.Context) ([]Location, error)
	// BroadcastLocationUpdate listens for location updates and sends them to a channel.
	BroadcastLocationUpdate(ctx context.Context, updates chan<- Location)
}

type locationService struct {
	redisClient *redis.Client
}

// NewLocationService returns a new instance of the locationService.
func NewLocationService(redisClient *redis.Client) LocationService {
	return &locationService{
		redisClient: redisClient,
	}
}

// UpdateLocation stores/updates the location in Redis with a 5 minute TTL.
func (ls *locationService) UpdateLocation(ctx context.Context, userID string, latitude, longitude float64) error {
	loc := Location{
		UserID:    userID,
		Latitude:  latitude,
		Longitude: longitude,
		Timestamp: time.Now(),
	}
	data, err := json.Marshal(loc)
	if err != nil {
		return err
	}
	key := "location:" + userID
	return ls.redisClient.Set(ctx, key, data, 5*time.Minute).Err()
}

// GetAllLocations fetches all keys matching "location:*" and returns their unmarshaled values.
func (ls *locationService) GetAllLocations(ctx context.Context) ([]Location, error) {
	keys, err := ls.redisClient.Keys(ctx, "location:*").Result()
	if err != nil {
		return nil, err
	}
	var locations []Location
	for _, key := range keys {
		data, err := ls.redisClient.Get(ctx, key).Result()
		if err != nil {
			continue
		}
		var loc Location
		if err := json.Unmarshal([]byte(data), &loc); err != nil {
			continue
		}
		locations = append(locations, loc)
	}
	return locations, nil
}

// BroadcastLocationUpdate sends location updates to a channel whenever a new location is updated.
func (ls *locationService) BroadcastLocationUpdate(ctx context.Context, updates chan<- Location) {
	pubsub := ls.redisClient.Subscribe(ctx, "__keyevent@0__:set")
	defer pubsub.Close()

	for {
		msg, err := pubsub.ReceiveMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return // Context canceled or deadline exceeded
			}
			continue // Ignore transient errors
		}

		if msg.Channel == "__keyevent@0__:set" && len(msg.Payload) > 9 && msg.Payload[:9] == "location:" {
			data, err := ls.redisClient.Get(ctx, msg.Payload).Result()
			if err != nil {
				continue
			}
			var loc Location
			if err := json.Unmarshal([]byte(data), &loc); err != nil {
				continue
			}
			select {
			case updates <- loc:
			case <-ctx.Done():
				return
			}
		}
	}
}
