package bookings

import (
	"context"
	"fmt"
	"math"

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

func (r *Repository) Create(ctx context.Context, booking *Booking, items []BookingItem) (string, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	bookingID := uuid.New().String()
	_, err = tx.Exec(ctx,
		`INSERT INTO bookings (id, organizer_id, itinerary_id, title, description, status, total_amount, discount_amount, escrow_amount, pricing_snapshot_id, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())`,
		bookingID, booking.OrganizerID, booking.ItineraryID, booking.Title, booking.Description,
		string(booking.Status), booking.TotalAmount, booking.DiscountAmount, booking.EscrowAmount, booking.PricingSnapshotID,
	)
	if err != nil {
		return "", fmt.Errorf("insert booking: %w", err)
	}

	for _, item := range items {
		itemID := uuid.New().String()
		subtotal := math.Round(item.UnitPrice*float64(item.Quantity)*100) / 100
		_, err = tx.Exec(ctx,
			`INSERT INTO booking_items (id, booking_id, item_type, item_name, description, unit_price, quantity, subtotal, category, created_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())`,
			itemID, bookingID, item.ItemType, item.ItemName, item.Description,
			item.UnitPrice, item.Quantity, subtotal, item.Category,
		)
		if err != nil {
			return "", fmt.Errorf("insert booking item: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("commit tx: %w", err)
	}
	return bookingID, nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (*Booking, error) {
	var b Booking
	err := r.pool.QueryRow(ctx,
		`SELECT id, organizer_id, itinerary_id, title, description, status, total_amount, discount_amount, escrow_amount, pricing_snapshot_id, created_at, updated_at
		 FROM bookings WHERE id = $1`, id,
	).Scan(
		&b.ID, &b.OrganizerID, &b.ItineraryID, &b.Title, &b.Description,
		&b.Status, &b.TotalAmount, &b.DiscountAmount, &b.EscrowAmount,
		&b.PricingSnapshotID, &b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get booking: %w", err)
	}
	return &b, nil
}

func (r *Repository) GetItems(ctx context.Context, bookingID string) ([]BookingItem, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, booking_id, item_type, item_name, description, unit_price, quantity, subtotal, category, created_at
		 FROM booking_items WHERE booking_id = $1 ORDER BY created_at`, bookingID)
	if err != nil {
		return nil, fmt.Errorf("query booking items: %w", err)
	}
	defer rows.Close()

	var items []BookingItem
	for rows.Next() {
		var item BookingItem
		err := rows.Scan(
			&item.ID, &item.BookingID, &item.ItemType, &item.ItemName, &item.Description,
			&item.UnitPrice, &item.Quantity, &item.Subtotal, &item.Category, &item.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan booking item: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) GetItemsTx(ctx context.Context, tx pgx.Tx, bookingID string) ([]BookingItem, error) {
	rows, err := tx.Query(ctx,
		`SELECT id, booking_id, item_type, item_name, description, unit_price, quantity, subtotal, category, created_at
		 FROM booking_items WHERE booking_id = $1 ORDER BY created_at`, bookingID)
	if err != nil {
		return nil, fmt.Errorf("query booking items tx: %w", err)
	}
	defer rows.Close()

	var items []BookingItem
	for rows.Next() {
		var item BookingItem
		err := rows.Scan(
			&item.ID, &item.BookingID, &item.ItemType, &item.ItemName, &item.Description,
			&item.UnitPrice, &item.Quantity, &item.Subtotal, &item.Category, &item.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan booking item tx: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) UpdateStatus(ctx context.Context, id string, status BookingStatus) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE bookings SET status = $1, updated_at = NOW() WHERE id = $2`,
		string(status), id,
	)
	if err != nil {
		return fmt.Errorf("update booking status: %w", err)
	}
	return nil
}

func (r *Repository) UpdateStatusTx(ctx context.Context, tx pgx.Tx, id string, status BookingStatus) error {
	_, err := tx.Exec(ctx,
		`UPDATE bookings SET status = $1, updated_at = NOW() WHERE id = $2`,
		string(status), id,
	)
	if err != nil {
		return fmt.Errorf("update booking status tx: %w", err)
	}
	return nil
}

func (r *Repository) UpdatePricingSnapshot(ctx context.Context, id, snapshotID string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE bookings SET pricing_snapshot_id = $1, updated_at = NOW() WHERE id = $2`,
		snapshotID, id,
	)
	if err != nil {
		return fmt.Errorf("update pricing snapshot: %w", err)
	}
	return nil
}

func (r *Repository) UpdateBookingAmountsTx(ctx context.Context, tx pgx.Tx, id string, totalAmount, discountAmount, escrowAmount float64, snapshotID string) error {
	_, err := tx.Exec(ctx,
		`UPDATE bookings SET total_amount = $1, discount_amount = $2, escrow_amount = $3, pricing_snapshot_id = $4, updated_at = NOW() WHERE id = $5`,
		totalAmount, discountAmount, escrowAmount, snapshotID, id,
	)
	if err != nil {
		return fmt.Errorf("update booking amounts tx: %w", err)
	}
	return nil
}

func (r *Repository) CreateEscrow(ctx context.Context, escrow *Escrow) error {
	id := uuid.New().String()
	_, err := r.pool.Exec(ctx,
		`INSERT INTO escrow_accounts (id, order_type, order_id, amount_held, amount_released, amount_refunded, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		id, escrow.OrderType, escrow.OrderID, escrow.AmountHeld, escrow.AmountReleased, escrow.AmountRefunded, string(escrow.Status),
	)
	if err != nil {
		return fmt.Errorf("create escrow: %w", err)
	}
	return nil
}

func (r *Repository) CreateEscrowTx(ctx context.Context, tx pgx.Tx, escrow *Escrow) error {
	id := uuid.New().String()
	_, err := tx.Exec(ctx,
		`INSERT INTO escrow_accounts (id, order_type, order_id, amount_held, amount_released, amount_refunded, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		id, escrow.OrderType, escrow.OrderID, escrow.AmountHeld, escrow.AmountReleased, escrow.AmountRefunded, string(escrow.Status),
	)
	if err != nil {
		return fmt.Errorf("create escrow tx: %w", err)
	}
	return nil
}

func (r *Repository) GetEscrow(ctx context.Context, orderType, orderID string) (*Escrow, error) {
	var e Escrow
	err := r.pool.QueryRow(ctx,
		`SELECT id, order_type, order_id, amount_held, amount_released, amount_refunded, status
		 FROM escrow_accounts WHERE order_type = $1 AND order_id = $2`, orderType, orderID,
	).Scan(&e.ID, &e.OrderType, &e.OrderID, &e.AmountHeld, &e.AmountReleased, &e.AmountRefunded, &e.Status)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get escrow: %w", err)
	}
	return &e, nil
}

func (r *Repository) UpdateEscrowStatus(ctx context.Context, id string, status EscrowStatus, released, refunded float64) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE escrow_accounts SET status = $1, amount_released = $2, amount_refunded = $3 WHERE id = $4`,
		string(status), released, refunded, id,
	)
	if err != nil {
		return fmt.Errorf("update escrow status: %w", err)
	}
	return nil
}

func (r *Repository) RecordTender(ctx context.Context, record *PaymentRecord) error {
	id := uuid.New().String()
	_, err := r.pool.Exec(ctx,
		`INSERT INTO payment_records (id, order_type, order_id, tender_type, amount, currency, reference_text, recorded_by, recorded_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())`,
		id, record.OrderType, record.OrderID, record.TenderType, record.Amount,
		record.Currency, record.ReferenceText, record.RecordedBy,
	)
	if err != nil {
		return fmt.Errorf("record tender: %w", err)
	}
	return nil
}

func (r *Repository) CreateCouponRedemption(ctx context.Context, redemption *CouponRedemption) error {
	id := uuid.New().String()
	_, err := r.pool.Exec(ctx,
		`INSERT INTO coupon_redemptions (id, coupon_id, user_id, booking_id, redemption_scope_key, discount_amount, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, NOW())`,
		id, redemption.CouponID, redemption.UserID, redemption.BookingID, redemption.RedemptionScopeKey, redemption.DiscountAmount,
	)
	if err != nil {
		return fmt.Errorf("create coupon redemption: %w", err)
	}
	return nil
}

func (r *Repository) CreateCouponRedemptionTx(ctx context.Context, tx pgx.Tx, redemption *CouponRedemption) error {
	id := uuid.New().String()
	_, err := tx.Exec(ctx,
		`INSERT INTO coupon_redemptions (id, coupon_id, user_id, booking_id, redemption_scope_key, discount_amount, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, NOW())`,
		id, redemption.CouponID, redemption.UserID, redemption.BookingID, redemption.RedemptionScopeKey, redemption.DiscountAmount,
	)
	if err != nil {
		return fmt.Errorf("create coupon redemption tx: %w", err)
	}
	return nil
}

func (r *Repository) GetByIDTx(ctx context.Context, tx pgx.Tx, id string) (*Booking, error) {
	var b Booking
	err := tx.QueryRow(ctx,
		`SELECT id, organizer_id, itinerary_id, title, description, status, total_amount, discount_amount, escrow_amount, pricing_snapshot_id, created_at, updated_at
		 FROM bookings WHERE id = $1 FOR UPDATE`, id,
	).Scan(
		&b.ID, &b.OrganizerID, &b.ItineraryID, &b.Title, &b.Description,
		&b.Status, &b.TotalAmount, &b.DiscountAmount, &b.EscrowAmount,
		&b.PricingSnapshotID, &b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get booking tx: %w", err)
	}
	return &b, nil
}
