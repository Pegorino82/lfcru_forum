package user

import (
	"context"
	"errors"
)

var ErrCannotBanSelf = errors.New("нельзя заблокировать самого себя")

// RepoInterface is the subset of Repo the Service depends on.
type RepoInterface interface {
	ListAll(ctx context.Context) ([]User, error)
	BanUser(ctx context.Context, id int64) error
	UnbanUser(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (*User, error)
}

type Service struct {
	repo RepoInterface
}

func NewService(repo RepoInterface) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListAll(ctx context.Context) ([]User, error) {
	return s.repo.ListAll(ctx)
}

// BanUser sets banned_at for the given user. Returns ErrCannotBanSelf if id == requestorID.
func (s *Service) BanUser(ctx context.Context, id, requestorID int64) error {
	if id == requestorID {
		return ErrCannotBanSelf
	}
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	return s.repo.BanUser(ctx, id)
}

// UnbanUser clears banned_at for the given user.
func (s *Service) UnbanUser(ctx context.Context, id int64) error {
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	return s.repo.UnbanUser(ctx, id)
}
