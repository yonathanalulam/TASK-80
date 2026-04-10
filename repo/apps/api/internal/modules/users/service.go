package users

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

type Service struct {
	repo   *Repository
	logger *zap.Logger
}

func NewService(repo *Repository, logger *zap.Logger) *Service {
	return &Service{repo: repo, logger: logger}
}

func (s *Service) GetUser(ctx context.Context, id string) (*UserResponse, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}

func (s *Service) UpdateProfile(ctx context.Context, userID, actorID string, actorRoles []string, req UpdateProfileRequest) error {
	if userID != actorID && !containsRole(actorRoles, "administrator") {
		return fmt.Errorf("forbidden")
	}
	return s.repo.UpdateProfile(ctx, userID, req)
}

func (s *Service) UpdatePreferences(ctx context.Context, userID, actorID string, actorRoles []string, req UpdatePreferencesRequest) error {
	if userID != actorID && !containsRole(actorRoles, "administrator") {
		return fmt.Errorf("forbidden")
	}
	return s.repo.UpdatePreferences(ctx, userID, req.Preferences)
}

func (s *Service) ListUsers(ctx context.Context, page, pageSize int, status string) ([]UserResponse, int, error) {
	return s.repo.List(ctx, page, pageSize, status)
}

func containsRole(roles []string, target string) bool {
	for _, r := range roles {
		if r == target {
			return true
		}
	}
	return false
}
