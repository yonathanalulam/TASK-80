package risk

import (
	"context"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"

	"travel-platform/apps/api/internal/common"
)

type Service struct {
	repo   *Repository
	engine *Engine
	logger *zap.Logger
}

func NewService(repo *Repository, logger *zap.Logger) *Service {
	return &Service{
		repo:   repo,
		engine: NewEngine(repo, logger),
		logger: logger,
	}
}

func (s *Service) GetRiskSummary(ctx context.Context, userID string) (*RiskSummaryDTO, error) {
	summary := &RiskSummaryDTO{UserID: userID}

	bl, err := s.repo.GetBlacklist(ctx, userID)
	if err != nil {
		return nil, common.NewInternalError("failed to check blacklist", err)
	}
	summary.IsBlacklisted = bl != nil

	rs, err := s.repo.GetRiskScore(ctx, userID)
	if err != nil {
		return nil, common.NewInternalError("failed to get risk score", err)
	}
	if rs != nil {
		summary.Score = rs.Score
	}

	throttles, err := s.repo.GetActiveThrottles(ctx, userID)
	if err != nil {
		return nil, common.NewInternalError("failed to get throttles", err)
	}
	summary.ActiveThrottles = make([]ThrottleActionDTO, 0, len(throttles))
	for _, t := range throttles {
		dto := ThrottleActionDTO{
			ID:         t.ID,
			ActionType: t.ActionType,
			Reason:     t.Reason,
			Active:     t.Active,
			CreatedAt:  t.CreatedAt,
		}
		if t.ExpiresAt != nil {
			dto.ExpiresAt = *t.ExpiresAt
		}
		summary.ActiveThrottles = append(summary.ActiveThrottles, dto)
	}

	events, err := s.repo.GetRecentEvents(ctx, userID, 20)
	if err != nil {
		s.logger.Error("failed to get recent events", zap.Error(err))
		events = nil
	}
	summary.RecentEvents = make([]RiskEventDTO, 0, len(events))
	for _, e := range events {
		summary.RecentEvents = append(summary.RecentEvents, RiskEventDTO{
			ID:          e.ID,
			EventType:   e.EventType,
			Description: e.Description,
			Severity:    e.Severity,
			CreatedAt:   e.CreatedAt,
		})
	}

	return summary, nil
}

func (s *Service) EvaluateAction(ctx context.Context, userID, actionType string) (*RiskDecision, error) {
	return s.engine.EvaluateAction(ctx, userID, actionType)
}

func (s *Service) RecordEvent(ctx context.Context, userID, eventType, description, severity string) error {
	metadata, _ := json.Marshal(map[string]string{"severity": severity})
	return s.repo.CreateRiskEvent(ctx, RiskEvent{
		UserID:       userID,
		EventType:    eventType,
		Description:  description,
		Severity:     severity,
		MetadataJSON: metadata,
	})
}

func (s *Service) BlacklistUser(ctx context.Context, targetID, adminID, reason string) error {
	if reason == "" {
		return common.NewBadRequestError("reason is required")
	}
	return s.repo.BlacklistUser(ctx, targetID, adminID, reason)
}

func (s *Service) UnblacklistUser(ctx context.Context, targetID, adminID string) error {
	_ = adminID
	return s.repo.UnblacklistUser(ctx, targetID)
}

func (s *Service) GetPendingApprovals(ctx context.Context) ([]AdminApprovalDTO, error) {
	approvals, err := s.repo.GetPendingApprovals(ctx)
	if err != nil {
		return nil, common.NewInternalError("failed to get pending approvals", err)
	}
	dtos := make([]AdminApprovalDTO, 0, len(approvals))
	for _, a := range approvals {
		dtos = append(dtos, AdminApprovalDTO{
			ID:              a.ID,
			UserID:          a.UserID,
			ActionType:      a.ActionType,
			ReferenceType:   a.ReferenceType,
			ReferenceID:     a.ReferenceID,
			Status:          a.Status,
			RequestedBy:     a.RequestedBy,
			ResolvedBy:      a.ResolvedBy,
			ResolutionNotes: a.ResolutionNotes,
			CreatedAt:       a.CreatedAt,
			ResolvedAt:      a.ResolvedAt,
		})
	}
	return dtos, nil
}

func (s *Service) ResolveApproval(ctx context.Context, approvalID, adminID, status, notes string) error {
	if status != "approved" && status != "rejected" {
		return fmt.Errorf("status must be 'approved' or 'rejected'")
	}
	return s.repo.ResolveApproval(ctx, approvalID, adminID, status, notes)
}
