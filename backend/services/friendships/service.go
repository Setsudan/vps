package friendships

import (
	"context"
	"errors"

	mfriend "launay-dot-one/models/friendships"
	"launay-dot-one/repositories"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type service struct {
	repo *repositories.FriendRequestRepository
	db   *gorm.DB
}

// NewService constructs the friendship service.
func NewService(repo *repositories.FriendRequestRepository, db *gorm.DB) Service {
	return &service{repo: repo, db: db}
}

func (s *service) SendRequest(ctx context.Context, fromID, toID string) error {
	if fromID == toID {
		return errors.New("cannot send friend request to yourself")
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1) check duplicates (unchanged)
		existing, err := s.repo.ListForUser(ctx, fromID)
		if err != nil {
			return err
		}
		for _, r := range existing {
			samePair := (r.RequesterID == fromID && r.ReceiverID == toID) ||
				(r.RequesterID == toID && r.ReceiverID == fromID)
			if samePair && (r.Status == string(mfriend.FriendPending) || r.Status == string(mfriend.FriendAccepted)) {
				return errors.New("friend request already exists")
			}
		}

		// 2) build the new request — ***populate ID here***
		fr := &mfriend.FriendRequest{
			ID:          uuid.NewString(),
			RequesterID: fromID,
			ReceiverID:  toID,
			Status:      string(mfriend.FriendPending),
		}
		return s.repo.Create(ctx, fr)
	})
}

func (s *service) RespondRequest(ctx context.Context, requestID string, accept bool) error {
	_, err := s.repo.Get(ctx, requestID)
	if err != nil {
		return err
	}
	newStatus := mfriend.FriendRejected
	if accept {
		newStatus = mfriend.FriendAccepted
	}
	return s.repo.UpdateStatus(ctx, requestID, newStatus)
}

func (s *service) ListRequests(ctx context.Context, userID string) ([]RequestDTO, error) {
	list, err := s.repo.ListForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]RequestDTO, len(list))
	for i, r := range list {
		out[i] = RequestDTO{
			ID:          r.ID,
			RequesterID: r.RequesterID,
			ReceiverID:  r.ReceiverID,
			Status:      mfriend.FriendStatus(r.Status),
		}
	}
	return out, nil
}

func (s *service) ListFriends(ctx context.Context, userID string) ([]FriendDTO, error) {
	frs, err := s.repo.ListFriends(ctx, userID)
	if err != nil {
		return nil, err
	}
	var out []FriendDTO
	for _, r := range frs {
		friendID := r.RequesterID
		if friendID == userID {
			friendID = r.ReceiverID
		}
		out = append(out, FriendDTO{UserID: friendID})
	}
	return out, nil
}
