package finance

import (
	"context"
	"errors"
	"fmt"
	"time"

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

func (r *Repository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.pool.Begin(ctx)
}
func (r *Repository) GetWallet(ctx context.Context, ownerID string, walletType WalletType) (*Wallet, error) {
	w := &Wallet{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, owner_id, wallet_type, balance, currency, created_at, updated_at
		 FROM wallets WHERE owner_id = $1 AND wallet_type = $2`, ownerID, walletType,
	).Scan(&w.ID, &w.OwnerID, &w.WalletType, &w.Balance, &w.Currency, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, common.NewNotFoundError("wallet")
		}
		return nil, fmt.Errorf("get wallet: %w", err)
	}
	return w, nil
}

func (r *Repository) GetWalletByID(ctx context.Context, walletID string) (*Wallet, error) {
	w := &Wallet{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, owner_id, wallet_type, balance, currency, created_at, updated_at
		 FROM wallets WHERE id = $1`, walletID,
	).Scan(&w.ID, &w.OwnerID, &w.WalletType, &w.Balance, &w.Currency, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, common.NewNotFoundError("wallet")
		}
		return nil, fmt.Errorf("get wallet by id: %w", err)
	}
	return w, nil
}

func (r *Repository) GetWalletTransactions(ctx context.Context, walletID string, page, pageSize int) ([]WalletTransaction, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var total int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM wallet_transactions WHERE wallet_id = $1`, walletID,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count wallet transactions: %w", err)
	}

	rows, err := r.pool.Query(ctx,
		`SELECT id, wallet_id, amount, direction, reference_type, reference_id, description, created_at
		 FROM wallet_transactions WHERE wallet_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		walletID, pageSize, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list wallet transactions: %w", err)
	}
	defer rows.Close()

	var items []WalletTransaction
	for rows.Next() {
		var t WalletTransaction
		if err := rows.Scan(&t.ID, &t.WalletID, &t.Amount, &t.Direction, &t.ReferenceType, &t.ReferenceID, &t.Description, &t.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan wallet transaction: %w", err)
		}
		items = append(items, t)
	}
	return items, total, rows.Err()
}

func (r *Repository) CreateWalletTransaction(ctx context.Context, tx pgx.Tx, walletID string, amount float64, direction Direction, refType, refID, desc string) error {
	_, err := tx.Exec(ctx,
		`INSERT INTO wallet_transactions (id, wallet_id, amount, direction, reference_type, reference_id, description, created_at)
		 VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, NOW())`,
		walletID, amount, direction, refType, refID, desc,
	)
	if err != nil {
		return fmt.Errorf("insert wallet transaction: %w", err)
	}
	return nil
}

func (r *Repository) UpdateWalletBalance(ctx context.Context, tx pgx.Tx, walletID string, delta float64) error {
	tag, err := tx.Exec(ctx,
		`UPDATE wallets SET balance = balance + $2, updated_at = NOW() WHERE id = $1`,
		walletID, delta,
	)
	if err != nil {
		return fmt.Errorf("update wallet balance: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return common.NewNotFoundError("wallet")
	}
	return nil
}
func (r *Repository) GetEscrow(ctx context.Context, orderType, orderID string) (*EscrowAccount, error) {
	e := &EscrowAccount{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, order_type, order_id, amount_held, amount_released, amount_refunded, status, created_at, updated_at
		 FROM escrow_accounts WHERE order_type = $1 AND order_id = $2`, orderType, orderID,
	).Scan(&e.ID, &e.OrderType, &e.OrderID, &e.AmountHeld, &e.AmountReleased, &e.AmountRefunded, &e.Status, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, common.NewNotFoundError("escrow account")
		}
		return nil, fmt.Errorf("get escrow: %w", err)
	}
	return e, nil
}

func (r *Repository) UpdateEscrowReleased(ctx context.Context, tx pgx.Tx, escrowID string, releaseAmount float64) error {
	tag, err := tx.Exec(ctx,
		`UPDATE escrow_accounts SET amount_released = amount_released + $2, status = 'partially_released', updated_at = NOW() WHERE id = $1`,
		escrowID, releaseAmount,
	)
	if err != nil {
		return fmt.Errorf("update escrow released: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return common.NewNotFoundError("escrow account")
	}
	return nil
}

func (r *Repository) UpdateEscrowRefunded(ctx context.Context, tx pgx.Tx, escrowID string, refundAmount float64) error {
	tag, err := tx.Exec(ctx,
		`UPDATE escrow_accounts SET amount_refunded = amount_refunded + $2, updated_at = NOW() WHERE id = $1`,
		escrowID, refundAmount,
	)
	if err != nil {
		return fmt.Errorf("update escrow refunded: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return common.NewNotFoundError("escrow account")
	}
	return nil
}

func (r *Repository) GetActiveEscrowsByOwner(ctx context.Context, ownerID string) ([]EscrowSummary, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT e.id, e.order_type, e.order_id, e.amount_held, e.amount_released, e.status, e.created_at
		 FROM escrow_accounts e
		 LEFT JOIN bookings b ON e.order_type = 'booking' AND e.order_id::text = b.id::text
		 LEFT JOIN purchase_orders p ON e.order_type = 'procurement' AND e.order_id::text = p.id::text
		 WHERE (b.organizer_id = $1 OR p.created_by = $1)
		   AND e.status IN ('held', 'partially_released')
		 ORDER BY e.created_at DESC`, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []EscrowSummary
	for rows.Next() {
		var es EscrowSummary
		var createdAt time.Time
		if err := rows.Scan(&es.ID, &es.OrderType, &es.OrderID, &es.AmountHeld, &es.AmountReleased, &es.Status, &createdAt); err != nil {
			return nil, err
		}
		es.CreatedAt = createdAt.Format(time.RFC3339)
		results = append(results, es)
	}
	if results == nil {
		results = []EscrowSummary{}
	}
	return results, rows.Err()
}
func (r *Repository) CreatePaymentRecord(ctx context.Context, p *PaymentRecord) (string, error) {
	var id string
	err := r.pool.QueryRow(ctx,
		`INSERT INTO payment_records (id, order_type, order_id, tender_type, amount, currency, reference_text, recorded_by, recorded_at)
		 VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, NOW()) RETURNING id`,
		p.OrderType, p.OrderID, p.TenderType, p.Amount, p.Currency, p.ReferenceText, p.RecordedBy,
	).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("insert payment record: %w", err)
	}
	return id, nil
}
func (r *Repository) CreateRefund(ctx context.Context, refund *Refund) (string, error) {
	var id string
	err := r.pool.QueryRow(ctx,
		`INSERT INTO refunds (id, order_type, order_id, refund_amount, refund_reason, created_by, status, created_at, updated_at)
		 VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, NOW(), NOW()) RETURNING id`,
		refund.OrderType, refund.OrderID, refund.RefundAmount, refund.RefundReason, refund.CreatedBy, refund.Status,
	).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("insert refund: %w", err)
	}
	return id, nil
}

func (r *Repository) CreateRefundItem(ctx context.Context, item *RefundItem) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO refund_items (id, refund_id, item_id, item_type, amount, created_at)
		 VALUES (gen_random_uuid(), $1, $2, $3, $4, NOW())`,
		item.RefundID, item.ItemID, item.ItemType, item.Amount,
	)
	if err != nil {
		return fmt.Errorf("insert refund item: %w", err)
	}
	return nil
}

func (r *Repository) GetRefund(ctx context.Context, id string) (*Refund, error) {
	ref := &Refund{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, order_type, order_id, refund_amount, refund_reason, created_by, approved_by, status, created_at, updated_at
		 FROM refunds WHERE id = $1`, id,
	).Scan(&ref.ID, &ref.OrderType, &ref.OrderID, &ref.RefundAmount, &ref.RefundReason, &ref.CreatedBy, &ref.ApprovedBy, &ref.Status, &ref.CreatedAt, &ref.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, common.NewNotFoundError("refund")
		}
		return nil, fmt.Errorf("get refund: %w", err)
	}
	return ref, nil
}
func (r *Repository) GetDailyWithdrawalTotal(ctx context.Context, courierID string, date time.Time) (float64, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	var total float64
	err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(request_amount), 0)
		 FROM withdrawal_requests
		 WHERE courier_id = $1 AND status != 'rejected' AND requested_at >= $2 AND requested_at < $3`,
		courierID, startOfDay, endOfDay,
	).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("get daily withdrawal total: %w", err)
	}
	return total, nil
}

func (r *Repository) CreateWithdrawalRequest(ctx context.Context, req *WithdrawalRequest) (string, error) {
	var id string
	err := r.pool.QueryRow(ctx,
		`INSERT INTO withdrawal_requests (id, courier_id, request_amount, status, requested_at, created_at, updated_at)
		 VALUES (gen_random_uuid(), $1, $2, $3, NOW(), NOW(), NOW()) RETURNING id`,
		req.CourierID, req.RequestAmount, req.Status,
	).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("insert withdrawal request: %w", err)
	}
	return id, nil
}

func (r *Repository) GetWithdrawalRequest(ctx context.Context, id string) (*WithdrawalRequest, error) {
	w := &WithdrawalRequest{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, courier_id, request_amount, status, requested_at, reviewed_by, approved_by, rejected_reason, settled_at, created_at, updated_at
		 FROM withdrawal_requests WHERE id = $1`, id,
	).Scan(&w.ID, &w.CourierID, &w.RequestAmount, &w.Status, &w.RequestedAt, &w.ReviewedBy, &w.ApprovedBy, &w.RejectedReason, &w.SettledAt, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, common.NewNotFoundError("withdrawal request")
		}
		return nil, fmt.Errorf("get withdrawal request: %w", err)
	}
	return w, nil
}

func (r *Repository) UpdateWithdrawalStatus(ctx context.Context, id, status string) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE withdrawal_requests SET status = $2, updated_at = NOW() WHERE id = $1`,
		id, status,
	)
	if err != nil {
		return fmt.Errorf("update withdrawal status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return common.NewNotFoundError("withdrawal request")
	}
	return nil
}

func (r *Repository) UpdateWithdrawalApproval(ctx context.Context, id, approverID, status string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE withdrawal_requests SET status = $2, approved_by = $3, reviewed_by = $3, settled_at = NOW(), updated_at = NOW() WHERE id = $1`,
		id, status, approverID,
	)
	if err != nil {
		return fmt.Errorf("update withdrawal approval: %w", err)
	}
	return nil
}

func (r *Repository) UpdateWithdrawalRejection(ctx context.Context, id, rejecterID, reason string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE withdrawal_requests SET status = 'rejected', reviewed_by = $2, rejected_reason = $3, updated_at = NOW() WHERE id = $1`,
		id, rejecterID, reason,
	)
	if err != nil {
		return fmt.Errorf("update withdrawal rejection: %w", err)
	}
	return nil
}

func (r *Repository) CreateWithdrawalDisbursement(ctx context.Context, withdrawalID string, amount float64) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO withdrawal_disbursements (id, withdrawal_id, amount, disbursed_at)
		 VALUES (gen_random_uuid(), $1, $2, NOW())`,
		withdrawalID, amount,
	)
	if err != nil {
		return fmt.Errorf("insert withdrawal disbursement: %w", err)
	}
	return nil
}
func (r *Repository) CreateReconciliationRun(ctx context.Context, run *ReconciliationRun) (string, error) {
	var id string
	err := r.pool.QueryRow(ctx,
		`INSERT INTO reconciliation_runs (id, run_date, status, summary_json, created_by, created_at)
		 VALUES (gen_random_uuid(), $1, $2, $3, $4, NOW()) RETURNING id`,
		run.RunDate, run.Status, run.SummaryJSON, run.CreatedBy,
	).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("insert reconciliation run: %w", err)
	}
	return id, nil
}

func (r *Repository) GetReconciliationSummary(ctx context.Context) (*ReconciliationReportDTO, error) {
	report := &ReconciliationReportDTO{}

	_ = r.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(amount), 0) FROM payment_records`,
	).Scan(&report.Inflows)

	_ = r.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(amount_held), 0) FROM escrow_accounts`,
	).Scan(&report.HeldInEscrow)

	_ = r.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(amount_released), 0) FROM escrow_accounts`,
	).Scan(&report.Released)

	_ = r.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(amount_refunded), 0) FROM escrow_accounts`,
	).Scan(&report.Refunded)

	_ = r.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(amount), 0) FROM withdrawal_disbursements`,
	).Scan(&report.Outflows)

	report.NetPayable = report.Inflows - report.Outflows - report.Refunded
	report.OpeningBalance = 0

	rows, err := r.pool.Query(ctx,
		`SELECT id, run_id, item_type, reference_id, expected_amount, actual_amount, difference, status, notes
		 FROM reconciliation_items WHERE status = 'unreconciled' ORDER BY item_type LIMIT 100`,
	)
	if err != nil {
		return report, nil
	}
	defer rows.Close()

	for rows.Next() {
		var item ReconciliationItem
		if err := rows.Scan(&item.ID, &item.RunID, &item.ItemType, &item.ReferenceID, &item.ExpectedAmount, &item.ActualAmount, &item.Difference, &item.Status, &item.Notes); err != nil {
			continue
		}
		report.UnreconciledItems = append(report.UnreconciledItems, item)
	}
	if report.UnreconciledItems == nil {
		report.UnreconciledItems = []ReconciliationItem{}
	}

	return report, nil
}
