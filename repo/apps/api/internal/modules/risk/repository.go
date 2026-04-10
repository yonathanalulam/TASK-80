package risk

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) CreateRiskEvent(ctx context.Context, event RiskEvent) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO risk_events (id, user_id, event_type, description, severity, metadata_json)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		uuid.New().String(), event.UserID, event.EventType, event.Description, event.Severity, event.MetadataJSON)
	return err
}

func (r *Repository) GetRecentEvents(ctx context.Context, userID string, limit int) ([]RiskEvent, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, event_type, description, severity, metadata_json, created_at
		 FROM risk_events WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2`,
		userID, limit)
	if err != nil {
		return nil, fmt.Errorf("get recent risk events: %w", err)
	}
	defer rows.Close()

	var events []RiskEvent
	for rows.Next() {
		var e RiskEvent
		if err := rows.Scan(&e.ID, &e.UserID, &e.EventType, &e.Description, &e.Severity, &e.MetadataJSON, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan risk event: %w", err)
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

func (r *Repository) GetRiskScore(ctx context.Context, userID string) (*RiskScore, error) {
	var rs RiskScore
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, score, factors_json, computed_at
		 FROM risk_scores WHERE user_id = $1 ORDER BY computed_at DESC LIMIT 1`, userID,
	).Scan(&rs.ID, &rs.UserID, &rs.Score, &rs.FactorsJSON, &rs.ComputedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &rs, nil
}

func (r *Repository) ComputeAndSaveRiskScore(ctx context.Context, userID string) (*RiskScore, error) {
	var highCount, mediumCount, lowCount int
	cutoff := time.Now().Add(-30 * 24 * time.Hour)

	_ = r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM risk_events WHERE user_id = $1 AND severity = 'high' AND created_at > $2`,
		userID, cutoff).Scan(&highCount)
	_ = r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM risk_events WHERE user_id = $1 AND severity = 'medium' AND created_at > $2`,
		userID, cutoff).Scan(&mediumCount)
	_ = r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM risk_events WHERE user_id = $1 AND severity = 'low' AND created_at > $2`,
		userID, cutoff).Scan(&lowCount)

	score := float64(highCount*10 + mediumCount*5 + lowCount)
	if score > 100 {
		score = 100
	}

	rs := &RiskScore{}
	err := r.pool.QueryRow(ctx,
		`INSERT INTO risk_scores (id, user_id, score, factors_json, computed_at)
		 VALUES ($1, $2, $3, $4, NOW()) RETURNING id, user_id, score, factors_json, computed_at`,
		uuid.New().String(), userID, score, fmt.Sprintf(`{"high":%d,"medium":%d,"low":%d}`, highCount, mediumCount, lowCount),
	).Scan(&rs.ID, &rs.UserID, &rs.Score, &rs.FactorsJSON, &rs.ComputedAt)
	if err != nil {
		return nil, fmt.Errorf("save risk score: %w", err)
	}
	return rs, nil
}

func (r *Repository) GetActiveThrottles(ctx context.Context, userID string) ([]ThrottleAction, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, action_type, reason, expires_at, active, created_by, created_at
		 FROM throttle_actions WHERE user_id = $1 AND active = true
		 AND (expires_at IS NULL OR expires_at > NOW())`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var actions []ThrottleAction
	for rows.Next() {
		var a ThrottleAction
		if err := rows.Scan(&a.ID, &a.UserID, &a.ActionType, &a.Reason, &a.ExpiresAt, &a.Active, &a.CreatedBy, &a.CreatedAt); err != nil {
			return nil, err
		}
		actions = append(actions, a)
	}
	return actions, rows.Err()
}

func (r *Repository) IsThrottled(ctx context.Context, userID, actionType string) (bool, *ThrottleAction, error) {
	a := &ThrottleAction{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, action_type, reason, expires_at, active, created_by, created_at
		 FROM throttle_actions
		 WHERE user_id = $1 AND (action_type = $2 OR action_type = '*') AND active = true
		 AND (expires_at IS NULL OR expires_at > NOW())
		 ORDER BY created_at DESC LIMIT 1`,
		userID, actionType,
	).Scan(&a.ID, &a.UserID, &a.ActionType, &a.Reason, &a.ExpiresAt, &a.Active, &a.CreatedBy, &a.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil, nil
		}
		return false, nil, err
	}
	return true, a, nil
}

func (r *Repository) CreateThrottle(ctx context.Context, action ThrottleAction) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO throttle_actions (id, user_id, action_type, reason, expires_at, active, created_by)
		 VALUES ($1, $2, $3, $4, $5, true, $6)`,
		uuid.New().String(), action.UserID, action.ActionType, action.Reason, action.ExpiresAt, action.CreatedBy)
	return err
}

func (r *Repository) CountCancellationsInWindow(ctx context.Context, userID string, hours int) (int, error) {
	var count int
	cutoff := time.Now().Add(-time.Duration(hours) * time.Hour)
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM risk_events
		 WHERE user_id = $1 AND event_type = 'cancellation'
		 AND created_at > $2`,
		userID, cutoff).Scan(&count)
	return count, err
}

func (r *Repository) CountRFQsInWindow(ctx context.Context, userID string, minutes int) (int, error) {
	var count int
	cutoff := time.Now().Add(-time.Duration(minutes) * time.Minute)
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM rfqs WHERE created_by = $1
		 AND created_at > $2`,
		userID, cutoff).Scan(&count)
	return count, err
}

func (r *Repository) CountHarassmentFlags(ctx context.Context, userID string) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM risk_events
		 WHERE user_id = $1 AND event_type = 'harassment_flag'`, userID).Scan(&count)
	return count, err
}

func (r *Repository) GetBlacklist(ctx context.Context, userID string) (*BlacklistRecord, error) {
	var br BlacklistRecord
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, reason, blacklisted_by, active, created_at, lifted_at
		 FROM blacklist_records WHERE user_id = $1 AND active = true LIMIT 1`, userID,
	).Scan(&br.ID, &br.UserID, &br.Reason, &br.BlacklistedBy, &br.Active, &br.CreatedAt, &br.LiftedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &br, nil
}

func (r *Repository) GetBlacklistRecord(ctx context.Context, userID string) (*BlacklistRecord, error) {
	return r.GetBlacklist(ctx, userID)
}

func (r *Repository) BlacklistUser(ctx context.Context, userID, adminID, reason string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO blacklist_records (id, user_id, reason, blacklisted_by, active)
		 VALUES ($1, $2, $3, $4, true)`,
		uuid.New().String(), userID, reason, adminID)
	return err
}

func (r *Repository) UnblacklistUser(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE blacklist_records SET active = false, lifted_at = NOW()
		 WHERE user_id = $1 AND active = true`, userID)
	return err
}

func (r *Repository) CreateAdminApproval(ctx context.Context, approval AdminApproval) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO admin_approvals (id, user_id, action_type, reference_type, reference_id, status, requested_by, created_at)
		 VALUES ($1, $2, $3, $4, $5, 'pending', $6, NOW())`,
		uuid.New().String(), approval.UserID, approval.ActionType, approval.ReferenceType, approval.ReferenceID, approval.RequestedBy)
	return err
}

func (r *Repository) GetPendingApprovals(ctx context.Context) ([]AdminApproval, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, action_type, reference_type, reference_id, status, requested_by, resolved_by, resolution_notes, created_at, resolved_at
		 FROM admin_approvals WHERE status = 'pending' ORDER BY created_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var approvals []AdminApproval
	for rows.Next() {
		var a AdminApproval
		if err := rows.Scan(&a.ID, &a.UserID, &a.ActionType, &a.ReferenceType, &a.ReferenceID, &a.Status, &a.RequestedBy, &a.ResolvedBy, &a.ResolutionNotes, &a.CreatedAt, &a.ResolvedAt); err != nil {
			return nil, err
		}
		approvals = append(approvals, a)
	}
	return approvals, rows.Err()
}

func (r *Repository) ResolveApproval(ctx context.Context, id, resolverID, status, notes string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE admin_approvals SET status = $2, resolved_by = $3, resolution_notes = $4, resolved_at = NOW()
		 WHERE id = $1`, id, status, resolverID, notes)
	return err
}
