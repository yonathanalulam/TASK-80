package pricing

import (
	"context"
	"encoding/json"
	"fmt"

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

func (r *Repository) GetActiveCouponsByCodes(ctx context.Context, codes []string) ([]Coupon, error) {
	if len(codes) == 0 {
		return nil, nil
	}

	rows, err := r.pool.Query(ctx,
		`SELECT id, code, name, discount_type, amount, min_spend, percent_off,
		        valid_from, valid_to, eligibility_json, stack_group, exclusive,
		        usage_limit_total, usage_limit_per_user, active, created_at, updated_at
		 FROM coupons
		 WHERE code = ANY($1) AND active = true`, codes)
	if err != nil {
		return nil, fmt.Errorf("query active coupons: %w", err)
	}
	defer rows.Close()

	var coupons []Coupon
	for rows.Next() {
		var c Coupon
		var eligJSON []byte
		err := rows.Scan(
			&c.ID, &c.Code, &c.Name, &c.DiscountType, &c.Amount, &c.MinSpend, &c.PercentOff,
			&c.ValidFrom, &c.ValidTo, &eligJSON, &c.StackGroup, &c.Exclusive,
			&c.UsageLimitTotal, &c.UsageLimitPerUser, &c.Active, &c.CreatedAt, &c.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan coupon: %w", err)
		}
		c.EligibilityJSON = json.RawMessage(eligJSON)
		coupons = append(coupons, c)
	}
	return coupons, rows.Err()
}

func (r *Repository) GetCouponRedemptionCount(ctx context.Context, couponID string) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM coupon_redemptions WHERE coupon_id = $1`, couponID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count coupon redemptions: %w", err)
	}
	return count, nil
}

func (r *Repository) GetUserCouponRedemptionCount(ctx context.Context, couponID, userID string) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM coupon_redemptions WHERE coupon_id = $1 AND user_id = $2`,
		couponID, userID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count user coupon redemptions: %w", err)
	}
	return count, nil
}

func (r *Repository) HasRedemption(ctx context.Context, couponID, userID, scopeKey string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM coupon_redemptions
			WHERE coupon_id = $1 AND user_id = $2 AND redemption_scope_key = $3
		)`, couponID, userID, scopeKey,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check redemption exists: %w", err)
	}
	return exists, nil
}

func (r *Repository) SavePricingSnapshot(ctx context.Context, bookingID *string, snapshotJSON []byte) (string, error) {
	id := uuid.New().String()
	_, err := r.pool.Exec(ctx,
		`INSERT INTO checkout_pricing_snapshots (id, booking_id, procurement_order_id, snapshot_json, created_at)
		 VALUES ($1, $2, NULL, $3, NOW())`,
		id, bookingID, snapshotJSON,
	)
	if err != nil {
		return "", fmt.Errorf("save pricing snapshot: %w", err)
	}
	return id, nil
}

func (r *Repository) SavePricingSnapshotTx(ctx context.Context, tx pgx.Tx, bookingID *string, snapshotJSON []byte) (string, error) {
	id := uuid.New().String()
	_, err := tx.Exec(ctx,
		`INSERT INTO checkout_pricing_snapshots (id, booking_id, procurement_order_id, snapshot_json, created_at)
		 VALUES ($1, $2, NULL, $3, NOW())`,
		id, bookingID, snapshotJSON,
	)
	if err != nil {
		return "", fmt.Errorf("save pricing snapshot tx: %w", err)
	}
	return id, nil
}

func (r *Repository) GetAllActiveCoupons(ctx context.Context) ([]Coupon, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, code, name, discount_type, amount, min_spend, percent_off,
		        valid_from, valid_to, eligibility_json, stack_group, exclusive,
		        usage_limit_total, usage_limit_per_user, active, created_at, updated_at
		 FROM coupons
		 WHERE active = true
		 ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("query all active coupons: %w", err)
	}
	defer rows.Close()

	var coupons []Coupon
	for rows.Next() {
		var c Coupon
		var eligJSON []byte
		err := rows.Scan(
			&c.ID, &c.Code, &c.Name, &c.DiscountType, &c.Amount, &c.MinSpend, &c.PercentOff,
			&c.ValidFrom, &c.ValidTo, &eligJSON, &c.StackGroup, &c.Exclusive,
			&c.UsageLimitTotal, &c.UsageLimitPerUser, &c.Active, &c.CreatedAt, &c.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan coupon: %w", err)
		}
		c.EligibilityJSON = json.RawMessage(eligJSON)
		coupons = append(coupons, c)
	}
	return coupons, rows.Err()
}
