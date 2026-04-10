package risk

import (
	"context"
	"encoding/json"
	"time"

	"go.uber.org/zap"
)

type Engine struct {
	repo   *Repository
	logger *zap.Logger
}

func NewEngine(repo *Repository, logger *zap.Logger) *Engine {
	return &Engine{repo: repo, logger: logger}
}

func (e *Engine) EvaluateAction(ctx context.Context, userID, actionType string) (*RiskDecision, error) {
	bl, err := e.repo.GetBlacklist(ctx, userID)
	if err != nil {
		return nil, err
	}
	if bl != nil {
		return &RiskDecision{Allowed: false, Reason: "account is blacklisted: " + bl.Reason}, nil
	}

	throttled, _, err := e.repo.IsThrottled(ctx, userID, actionType)
	if err != nil {
		return nil, err
	}
	if throttled {
		return &RiskDecision{Allowed: false, Reason: "action is currently throttled"}, nil
	}

	cancellations, err := e.repo.CountCancellationsInWindow(ctx, userID, 24)
	if err != nil {
		e.logger.Error("failed to count cancellations", zap.Error(err))
	} else if cancellations > 8 {
		expires := time.Now().Add(6 * time.Hour)
		_ = e.repo.CreateThrottle(ctx, ThrottleAction{
			UserID:     userID,
			ActionType: actionType,
			Reason:     "exceeded 8 cancellations in 24 hours",
			ExpiresAt:  &expires,
		})
		_ = e.repo.CreateAdminApproval(ctx, AdminApproval{
			UserID:        &userID,
			ActionType:    actionType,
			ReferenceType: "cancellation_throttle",
			ReferenceID:   userID,
			RequestedBy:   "system",
		})
		return &RiskDecision{Allowed: false, RequireApproval: true, Reason: "too many cancellations in 24 hours"}, nil
	}

	if actionType == "create_rfq" {
		rfqCount, err := e.repo.CountRFQsInWindow(ctx, userID, 10)
		if err != nil {
			e.logger.Error("failed to count RFQs", zap.Error(err))
		} else if rfqCount >= 20 {
			expires := time.Now().Add(1 * time.Hour)
			_ = e.repo.CreateThrottle(ctx, ThrottleAction{
				UserID:     userID,
				ActionType: "create_rfq",
				Reason:     "exceeded 20 RFQs in 10 minutes",
				ExpiresAt:  &expires,
			})
			return &RiskDecision{Allowed: false, Reason: "too many RFQ requests"}, nil
		}
	}

	flags, err := e.repo.CountHarassmentFlags(ctx, userID)
	if err != nil {
		e.logger.Error("failed to count harassment flags", zap.Error(err))
	} else if flags >= 3 {
		expires := time.Now().Add(24 * time.Hour)
		_ = e.repo.CreateThrottle(ctx, ThrottleAction{
			UserID:     userID,
			ActionType: "*",
			Reason:     "multiple harassment flags",
			ExpiresAt:  &expires,
		})
		return &RiskDecision{Allowed: false, Reason: "account frozen due to multiple harassment reports"}, nil
	}

	return &RiskDecision{Allowed: true}, nil
}

func (e *Engine) RecordAndEvaluate(ctx context.Context, userID, eventType, description string, metadata map[string]interface{}) (*RiskDecision, error) {
	metadataJSON, _ := json.Marshal(metadata)

	err := e.repo.CreateRiskEvent(ctx, RiskEvent{
		UserID:       userID,
		EventType:    eventType,
		Description:  description,
		Severity:     determineSeverity(eventType),
		MetadataJSON: metadataJSON,
	})
	if err != nil {
		return nil, err
	}

	_, err = e.repo.ComputeAndSaveRiskScore(ctx, userID)
	if err != nil {
		e.logger.Error("failed to compute risk score", zap.Error(err))
	}

	return e.EvaluateAction(ctx, userID, eventType)
}

func determineSeverity(eventType string) string {
	switch eventType {
	case "harassment_flag":
		return "high"
	case "cancellation":
		return "medium"
	case "rfq_creation":
		return "low"
	default:
		return "low"
	}
}
