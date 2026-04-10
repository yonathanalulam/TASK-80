package reviews

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"travel-platform/apps/api/internal/common"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}
func (r *Repository) CreateReview(ctx context.Context, rev *Review) (string, error) {
	var id string
	err := r.pool.QueryRow(ctx,
		`INSERT INTO reviews (id, reviewer_id, subject_id, order_type, order_id, overall_rating, comment, editable_until, created_at, updated_at)
		 VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		 RETURNING id`,
		rev.ReviewerID, rev.SubjectID, rev.OrderType, rev.OrderID,
		rev.OverallRating, rev.Comment, rev.EditableUntil,
	).Scan(&id)
	if err != nil {
		if isDuplicateKey(err) {
			return "", common.NewConflictError("you have already reviewed this transaction")
		}
		return "", fmt.Errorf("insert review: %w", err)
	}
	return id, nil
}

func (r *Repository) GetReviewsBySubject(ctx context.Context, subjectID string, page, pageSize int) ([]Review, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var total int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM reviews WHERE subject_id = $1`, subjectID,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count reviews: %w", err)
	}

	rows, err := r.pool.Query(ctx,
		`SELECT id, reviewer_id, subject_id, order_type, order_id, overall_rating, comment, editable_until, created_at, updated_at
		 FROM reviews WHERE subject_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		subjectID, pageSize, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list reviews: %w", err)
	}
	defer rows.Close()

	var items []Review
	for rows.Next() {
		var rev Review
		if err := rows.Scan(
			&rev.ID, &rev.ReviewerID, &rev.SubjectID, &rev.OrderType, &rev.OrderID,
			&rev.OverallRating, &rev.Comment, &rev.EditableUntil, &rev.CreatedAt, &rev.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan review: %w", err)
		}
		items = append(items, rev)
	}
	return items, total, rows.Err()
}
func (r *Repository) GetDimensionByName(ctx context.Context, name string) (*ReviewDimension, error) {
	d := &ReviewDimension{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, label, active FROM review_dimensions WHERE name = $1`, name,
	).Scan(&d.ID, &d.Name, &d.Label, &d.Active)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, common.NewNotFoundError("review dimension")
		}
		return nil, fmt.Errorf("get dimension: %w", err)
	}
	return d, nil
}

func (r *Repository) CreateReviewScore(ctx context.Context, reviewID, dimensionID string, score float64) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO review_scores (id, review_id, dimension_id, score)
		 VALUES (gen_random_uuid(), $1, $2, $3)`,
		reviewID, dimensionID, score,
	)
	if err != nil {
		return fmt.Errorf("insert review score: %w", err)
	}
	return nil
}

func (r *Repository) GetScoresForReview(ctx context.Context, reviewID string) ([]ScoreDetail, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT rd.name, rs.score
		 FROM review_scores rs
		 JOIN review_dimensions rd ON rd.id = rs.dimension_id
		 WHERE rs.review_id = $1`, reviewID,
	)
	if err != nil {
		return nil, fmt.Errorf("get review scores: %w", err)
	}
	defer rows.Close()

	var items []ScoreDetail
	for rows.Next() {
		var s ScoreDetail
		if err := rows.Scan(&s.DimensionName, &s.Score); err != nil {
			return nil, fmt.Errorf("scan review score: %w", err)
		}
		items = append(items, s)
	}
	return items, rows.Err()
}
func (r *Repository) GetAllCreditTiers(ctx context.Context) ([]CreditTier, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, tier_name, min_transactions, min_avg_rating, max_violations, description
		 FROM credit_tiers ORDER BY min_transactions DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("get credit tiers: %w", err)
	}
	defer rows.Close()

	var items []CreditTier
	for rows.Next() {
		var ct CreditTier
		if err := rows.Scan(&ct.ID, &ct.TierName, &ct.MinTransactions, &ct.MinAvgRating, &ct.MaxViolations, &ct.Description); err != nil {
			return nil, fmt.Errorf("scan credit tier: %w", err)
		}
		items = append(items, ct)
	}
	return items, rows.Err()
}

func (r *Repository) GetAvgRatingForSubject(ctx context.Context, userID string) (float64, int, error) {
	var avg float64
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(AVG(overall_rating), 0), COUNT(*) FROM reviews WHERE subject_id = $1`, userID,
	).Scan(&avg, &count)
	if err != nil {
		return 0, 0, fmt.Errorf("get avg rating: %w", err)
	}
	return avg, count, nil
}

func (r *Repository) GetViolationCount(ctx context.Context, userID string) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM violation_records WHERE user_id = $1`, userID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count violations: %w", err)
	}
	return count, nil
}

func (r *Repository) CreateCreditSnapshot(ctx context.Context, snap *UserCreditSnapshot) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO user_credit_snapshots (id, user_id, tier, total_transactions, avg_rating, violation_count, computed_at)
		 VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, NOW())
		 RETURNING id, computed_at`,
		snap.UserID, snap.Tier, snap.TotalTransactions, snap.AvgRating, snap.ViolationCount,
	).Scan(&snap.ID, &snap.ComputedAt)
}

func (r *Repository) GetLatestCreditSnapshot(ctx context.Context, userID string) (*UserCreditSnapshot, error) {
	snap := &UserCreditSnapshot{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, tier, total_transactions, avg_rating, violation_count, computed_at
		 FROM user_credit_snapshots WHERE user_id = $1 ORDER BY computed_at DESC LIMIT 1`, userID,
	).Scan(&snap.ID, &snap.UserID, &snap.Tier, &snap.TotalTransactions, &snap.AvgRating, &snap.ViolationCount, &snap.ComputedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, common.NewNotFoundError("credit snapshot")
		}
		return nil, fmt.Errorf("get credit snapshot: %w", err)
	}
	return snap, nil
}
func (r *Repository) CreateViolation(ctx context.Context, v *ViolationRecord) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO violation_records (id, user_id, violation_type, description, severity, recorded_by, created_at)
		 VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, NOW())`,
		v.UserID, v.ViolationType, v.Description, v.Severity, v.RecordedBy,
	)
	if err != nil {
		return fmt.Errorf("insert violation: %w", err)
	}
	return nil
}
func (r *Repository) CreateNoShow(ctx context.Context, ns *NoShowRecord) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO no_show_records (id, user_id, order_type, order_id, recorded_by, created_at)
		 VALUES (gen_random_uuid(), $1, $2, $3, $4, NOW())`,
		ns.UserID, ns.OrderType, ns.OrderID, ns.RecordedBy,
	)
	if err != nil {
		return fmt.Errorf("insert no-show: %w", err)
	}
	return nil
}

func (r *Repository) GetNoShowCount(ctx context.Context, userID string) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM no_show_records WHERE user_id = $1`, userID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count no-shows: %w", err)
	}
	return count, nil
}
func (r *Repository) CreateHarassmentFlag(ctx context.Context, hf *HarassmentFlag) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO harassment_flags (id, reporter_id, subject_id, description, evidence_file_id, status, created_at, updated_at)
		 VALUES (gen_random_uuid(), $1, $2, $3, $4, 'pending', NOW(), NOW())
		 RETURNING id, created_at, updated_at`,
		hf.ReporterID, hf.SubjectID, hf.Description, hf.EvidenceFileID,
	).Scan(&hf.ID, &hf.CreatedAt, &hf.UpdatedAt)
}

func (r *Repository) GetHarassmentFlagCount(ctx context.Context, userID string) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM harassment_flags WHERE subject_id = $1`, userID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count harassment flags: %w", err)
	}
	return count, nil
}
func (r *Repository) IsBlacklisted(ctx context.Context, userID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM blacklist_records WHERE user_id = $1 AND active = TRUE)`, userID,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check blacklist: %w", err)
	}
	return exists, nil
}
func isDuplicateKey(err error) bool {
	if err == nil {
		return false
	}
	type pgErr interface {
		SQLState() string
	}
	var pe pgErr
	if errors.As(err, &pe) {
		return pe.SQLState() == "23505"
	}
	return false
}
