package admin

import (
	"context"

	"go.uber.org/zap"
)

type Service struct {
	repo   *Repository
	logger *zap.Logger
}

func NewService(repo *Repository, logger *zap.Logger) *Service {
	return &Service{repo: repo, logger: logger}
}

func (s *Service) GetAuditLogs(ctx context.Context, filters AuditLogFilters) ([]AuditLog, int, error) {
	return s.repo.GetAuditLogs(ctx, filters)
}

func (s *Service) GetConfig(_ context.Context) map[string]interface{} {
	return map[string]interface{}{
		"dnd_default_start":       "21:00",
		"dnd_default_end":         "08:00",
		"courier_daily_cap":       2500.00,
		"refund_minimum_unit":     1.00,
		"download_token_ttl_min":  5,
		"max_cancellations_24h":   8,
		"max_rfqs_10min":          20,
		"coupon_max_threshold":    1,
		"coupon_max_percentage":   1,
		"new_user_gift_exclusive": true,
	}
}
