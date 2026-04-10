package procurement

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
func (r *Repository) CreateRFQ(ctx context.Context, rfq *RFQ) (string, error) {
	var id string
	err := r.pool.QueryRow(ctx,
		`INSERT INTO rfqs (id, created_by, title, description, deadline, status, created_at, updated_at)
		 VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, NOW(), NOW()) RETURNING id`,
		rfq.CreatedBy, rfq.Title, rfq.Description, rfq.Deadline, rfq.Status,
	).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("insert rfq: %w", err)
	}
	return id, nil
}

func (r *Repository) GetRFQ(ctx context.Context, id string) (*RFQ, error) {
	rfq := &RFQ{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, created_by, title, description, deadline, status, created_at, updated_at
		 FROM rfqs WHERE id = $1`, id,
	).Scan(&rfq.ID, &rfq.CreatedBy, &rfq.Title, &rfq.Description, &rfq.Deadline, &rfq.Status, &rfq.CreatedAt, &rfq.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, common.NewNotFoundError("rfq")
		}
		return nil, fmt.Errorf("get rfq: %w", err)
	}
	return rfq, nil
}

func (r *Repository) UpdateRFQStatus(ctx context.Context, id string, status RFQStatus) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE rfqs SET status = $2, updated_at = NOW() WHERE id = $1`, id, status,
	)
	if err != nil {
		return fmt.Errorf("update rfq status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return common.NewNotFoundError("rfq")
	}
	return nil
}
func (r *Repository) CreateRFQItem(ctx context.Context, item *RFQItem) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO rfq_items (id, rfq_id, item_name, specifications, quantity, unit, sort_order, created_at)
		 VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, NOW())`,
		item.RFQID, item.ItemName, item.Specifications, item.Quantity, item.Unit, item.SortOrder,
	)
	if err != nil {
		return fmt.Errorf("insert rfq item: %w", err)
	}
	return nil
}

func (r *Repository) GetRFQItems(ctx context.Context, rfqID string) ([]RFQItem, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, rfq_id, item_name, specifications, quantity, unit, sort_order, created_at
		 FROM rfq_items WHERE rfq_id = $1 ORDER BY sort_order`, rfqID,
	)
	if err != nil {
		return nil, fmt.Errorf("get rfq items: %w", err)
	}
	defer rows.Close()

	var items []RFQItem
	for rows.Next() {
		var item RFQItem
		if err := rows.Scan(&item.ID, &item.RFQID, &item.ItemName, &item.Specifications, &item.Quantity, &item.Unit, &item.SortOrder, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan rfq item: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
func (r *Repository) InviteSupplier(ctx context.Context, rfqID, supplierID string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO rfq_suppliers (id, rfq_id, supplier_id, invited_at)
		 VALUES (gen_random_uuid(), $1, $2, NOW())`,
		rfqID, supplierID,
	)
	if err != nil {
		return fmt.Errorf("invite supplier: %w", err)
	}
	return nil
}

func (r *Repository) IsSupplierInvited(ctx context.Context, rfqID, supplierID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM rfq_suppliers WHERE rfq_id = $1 AND supplier_id = $2)`,
		rfqID, supplierID,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check supplier invitation: %w", err)
	}
	return exists, nil
}
func (r *Repository) CreateQuote(ctx context.Context, quote *RFQQuote) (string, error) {
	var id string
	err := r.pool.QueryRow(ctx,
		`INSERT INTO rfq_quotes (id, rfq_id, supplier_id, total_amount, lead_time_days, notes, submitted_at, created_at)
		 VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, NOW(), NOW()) RETURNING id`,
		quote.RFQID, quote.SupplierID, quote.TotalAmount, quote.LeadTimeDays, quote.Notes,
	).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("insert quote: %w", err)
	}
	return id, nil
}

func (r *Repository) CreateQuoteItem(ctx context.Context, item *RFQQuoteItem) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO rfq_quote_items (id, quote_id, rfq_item_id, unit_price, quantity, subtotal, notes)
		 VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6)`,
		item.QuoteID, item.RFQItemID, item.UnitPrice, item.Quantity, item.Subtotal, item.Notes,
	)
	if err != nil {
		return fmt.Errorf("insert quote item: %w", err)
	}
	return nil
}

func (r *Repository) GetQuotesByRFQ(ctx context.Context, rfqID string) ([]RFQQuote, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, rfq_id, supplier_id, total_amount, lead_time_days, notes, submitted_at, created_at
		 FROM rfq_quotes WHERE rfq_id = $1 ORDER BY total_amount`, rfqID,
	)
	if err != nil {
		return nil, fmt.Errorf("get quotes by rfq: %w", err)
	}
	defer rows.Close()

	var quotes []RFQQuote
	for rows.Next() {
		var q RFQQuote
		if err := rows.Scan(&q.ID, &q.RFQID, &q.SupplierID, &q.TotalAmount, &q.LeadTimeDays, &q.Notes, &q.SubmittedAt, &q.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan quote: %w", err)
		}
		quotes = append(quotes, q)
	}
	return quotes, rows.Err()
}

func (r *Repository) GetQuote(ctx context.Context, id string) (*RFQQuote, error) {
	q := &RFQQuote{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, rfq_id, supplier_id, total_amount, lead_time_days, notes, submitted_at, created_at
		 FROM rfq_quotes WHERE id = $1`, id,
	).Scan(&q.ID, &q.RFQID, &q.SupplierID, &q.TotalAmount, &q.LeadTimeDays, &q.Notes, &q.SubmittedAt, &q.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, common.NewNotFoundError("quote")
		}
		return nil, fmt.Errorf("get quote: %w", err)
	}
	return q, nil
}

func (r *Repository) GetQuoteItems(ctx context.Context, quoteID string) ([]RFQQuoteItem, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, quote_id, rfq_item_id, unit_price, quantity, subtotal, notes
		 FROM rfq_quote_items WHERE quote_id = $1`, quoteID,
	)
	if err != nil {
		return nil, fmt.Errorf("get quote items: %w", err)
	}
	defer rows.Close()

	var items []RFQQuoteItem
	for rows.Next() {
		var item RFQQuoteItem
		if err := rows.Scan(&item.ID, &item.QuoteID, &item.RFQItemID, &item.UnitPrice, &item.Quantity, &item.Subtotal, &item.Notes); err != nil {
			return nil, fmt.Errorf("scan quote item: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
func (r *Repository) CreatePO(ctx context.Context, po *PurchaseOrder) (string, error) {
	var id string
	err := r.pool.QueryRow(ctx,
		`INSERT INTO purchase_orders (id, rfq_id, quote_id, supplier_id, created_by, po_number, promised_date, status, total_amount, created_at, updated_at)
		 VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW()) RETURNING id`,
		po.RFQID, po.QuoteID, po.SupplierID, po.CreatedBy, po.PONumber, po.PromisedDate, po.Status, po.TotalAmount,
	).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("insert purchase order: %w", err)
	}
	return id, nil
}

func (r *Repository) CreatePOItem(ctx context.Context, item *POItem) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO po_items (id, po_id, item_name, specifications, unit_price, quantity, subtotal, created_at)
		 VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, NOW())`,
		item.POID, item.ItemName, item.Specifications, item.UnitPrice, item.Quantity, item.Subtotal,
	)
	if err != nil {
		return fmt.Errorf("insert po item: %w", err)
	}
	return nil
}

func (r *Repository) GetPO(ctx context.Context, id string) (*PurchaseOrder, error) {
	po := &PurchaseOrder{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, rfq_id, quote_id, supplier_id, created_by, po_number, promised_date, status, total_amount, created_at, updated_at
		 FROM purchase_orders WHERE id = $1`, id,
	).Scan(&po.ID, &po.RFQID, &po.QuoteID, &po.SupplierID, &po.CreatedBy, &po.PONumber, &po.PromisedDate, &po.Status, &po.TotalAmount, &po.CreatedAt, &po.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, common.NewNotFoundError("purchase order")
		}
		return nil, fmt.Errorf("get purchase order: %w", err)
	}
	return po, nil
}

func (r *Repository) GetPOItems(ctx context.Context, poID string) ([]POItem, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, po_id, item_name, specifications, unit_price, quantity, subtotal, created_at
		 FROM po_items WHERE po_id = $1`, poID,
	)
	if err != nil {
		return nil, fmt.Errorf("get po items: %w", err)
	}
	defer rows.Close()

	var items []POItem
	for rows.Next() {
		var item POItem
		if err := rows.Scan(&item.ID, &item.POID, &item.ItemName, &item.Specifications, &item.UnitPrice, &item.Quantity, &item.Subtotal, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan po item: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) UpdatePOStatus(ctx context.Context, id string, status POStatus) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE purchase_orders SET status = $2, updated_at = NOW() WHERE id = $1`, id, status,
	)
	if err != nil {
		return fmt.Errorf("update po status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return common.NewNotFoundError("purchase order")
	}
	return nil
}

func (r *Repository) GetNextPOSequence(ctx context.Context, year int) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM purchase_orders WHERE EXTRACT(YEAR FROM created_at) = $1`, year,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("get next po sequence: %w", err)
	}
	return count + 1, nil
}
func (r *Repository) CreateDelivery(ctx context.Context, d *Delivery) (string, error) {
	var id string
	err := r.pool.QueryRow(ctx,
		`INSERT INTO deliveries (id, po_id, courier_id, received_by, delivery_date, notes, status, created_at, updated_at)
		 VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, NOW(), NOW()) RETURNING id`,
		d.POID, d.CourierID, d.ReceivedBy, d.DeliveryDate, d.Notes, d.Status,
	).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("insert delivery: %w", err)
	}
	return id, nil
}

func (r *Repository) CreateDeliveryItem(ctx context.Context, item *DeliveryItem) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO delivery_items (id, delivery_id, po_item_id, quantity_delivered, quantity_accepted, quantity_rejected, created_at)
		 VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, NOW())`,
		item.DeliveryID, item.POItemID, item.QuantityDelivered, item.QuantityAccepted, item.QuantityRejected,
	)
	if err != nil {
		return fmt.Errorf("insert delivery item: %w", err)
	}
	return nil
}

func (r *Repository) GetDelivery(ctx context.Context, id string) (*Delivery, error) {
	d := &Delivery{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, po_id, courier_id, received_by, delivery_date, notes, status, created_at, updated_at
		 FROM deliveries WHERE id = $1`, id,
	).Scan(&d.ID, &d.POID, &d.CourierID, &d.ReceivedBy, &d.DeliveryDate, &d.Notes, &d.Status, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, common.NewNotFoundError("delivery")
		}
		return nil, fmt.Errorf("get delivery: %w", err)
	}
	return d, nil
}

func (r *Repository) GetDeliveryItems(ctx context.Context, deliveryID string) ([]DeliveryItem, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, delivery_id, po_item_id, quantity_delivered, quantity_accepted, quantity_rejected, created_at
		 FROM delivery_items WHERE delivery_id = $1`, deliveryID,
	)
	if err != nil {
		return nil, fmt.Errorf("get delivery items: %w", err)
	}
	defer rows.Close()

	var items []DeliveryItem
	for rows.Next() {
		var item DeliveryItem
		if err := rows.Scan(&item.ID, &item.DeliveryID, &item.POItemID, &item.QuantityDelivered, &item.QuantityAccepted, &item.QuantityRejected, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan delivery item: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
func (r *Repository) CreateInspection(ctx context.Context, insp *QualityInspection) (string, error) {
	var id string
	err := r.pool.QueryRow(ctx,
		`INSERT INTO quality_inspections (id, delivery_id, po_id, inspector_id, status, notes, inspected_at, created_at, updated_at)
		 VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, NOW(), NOW(), NOW()) RETURNING id`,
		insp.DeliveryID, insp.POID, insp.InspectorID, insp.Status, insp.Notes,
	).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("insert inspection: %w", err)
	}
	return id, nil
}
func (r *Repository) CreateDiscrepancy(ctx context.Context, d *DiscrepancyTicket) (string, error) {
	var id string
	err := r.pool.QueryRow(ctx,
		`INSERT INTO discrepancy_tickets (id, po_id, delivery_id, inspection_id, discrepancy_type, description, notes, status, created_by, created_at, updated_at)
		 VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, 'open', $7, NOW(), NOW()) RETURNING id`,
		d.POID, d.DeliveryID, d.InspectionID, d.DiscrepancyType, d.Description, d.Notes, d.CreatedBy,
	).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("insert discrepancy: %w", err)
	}
	return id, nil
}
func (r *Repository) CreateExceptionCase(ctx context.Context, ec *ExceptionCase) (string, error) {
	var id string
	err := r.pool.QueryRow(ctx,
		`INSERT INTO exception_cases (id, reference_type, reference_id, status, opened_reason, opened_at, created_at, updated_at)
		 VALUES (gen_random_uuid(), $1, $2, $3, $4, NOW(), NOW(), NOW()) RETURNING id`,
		ec.ReferenceType, ec.ReferenceID, ec.Status, ec.OpenedReason,
	).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("insert exception case: %w", err)
	}
	return id, nil
}

func (r *Repository) GetExceptionCase(ctx context.Context, id string) (*ExceptionCase, error) {
	ec := &ExceptionCase{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, reference_type, reference_id, status, opened_reason, opened_at, closed_at, created_at, updated_at
		 FROM exception_cases WHERE id = $1`, id,
	).Scan(&ec.ID, &ec.ReferenceType, &ec.ReferenceID, &ec.Status, &ec.OpenedReason, &ec.OpenedAt, &ec.ClosedAt, &ec.CreatedAt, &ec.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, common.NewNotFoundError("exception case")
		}
		return nil, fmt.Errorf("get exception case: %w", err)
	}
	return ec, nil
}

func (r *Repository) CloseExceptionCase(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE exception_cases SET status = 'closed', closed_at = NOW(), updated_at = NOW() WHERE id = $1`, id,
	)
	if err != nil {
		return fmt.Errorf("close exception case: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return common.NewNotFoundError("exception case")
	}
	return nil
}
func (r *Repository) CreateWaiver(ctx context.Context, w *WaiverRecord) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO waiver_records (id, exception_case_id, approved_by, waiver_reason, created_at)
		 VALUES (gen_random_uuid(), $1, $2, $3, NOW())`,
		w.ExceptionCaseID, w.ApprovedBy, w.WaiverReason,
	)
	if err != nil {
		return fmt.Errorf("insert waiver: %w", err)
	}
	return nil
}

func (r *Repository) GetWaiversByException(ctx context.Context, exceptionID string) ([]WaiverRecord, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, exception_case_id, approved_by, waiver_reason, created_at
		 FROM waiver_records WHERE exception_case_id = $1 ORDER BY created_at`, exceptionID,
	)
	if err != nil {
		return nil, fmt.Errorf("get waivers: %w", err)
	}
	defer rows.Close()

	var waivers []WaiverRecord
	for rows.Next() {
		var w WaiverRecord
		if err := rows.Scan(&w.ID, &w.ExceptionCaseID, &w.ApprovedBy, &w.WaiverReason, &w.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan waiver: %w", err)
		}
		waivers = append(waivers, w)
	}
	return waivers, rows.Err()
}

func (r *Repository) CountWaiversByException(ctx context.Context, exceptionID string) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM waiver_records WHERE exception_case_id = $1`, exceptionID,
	).Scan(&count)
	return count, err
}
func (r *Repository) CreateSettlementAdjustment(ctx context.Context, sa *SettlementAdjustment) (string, error) {
	var id string
	err := r.pool.QueryRow(ctx,
		`INSERT INTO settlement_adjustments (id, exception_case_id, amount, direction, reason, approved_by, journal_entry_id, created_at)
		 VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, NOW()) RETURNING id`,
		sa.ExceptionCaseID, sa.Amount, sa.Direction, sa.Reason, sa.ApprovedBy, sa.JournalEntryID,
	).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("insert settlement adjustment: %w", err)
	}
	return id, nil
}

func (r *Repository) CountAdjustmentsByException(ctx context.Context, exceptionID string) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM settlement_adjustments WHERE exception_case_id = $1`, exceptionID,
	).Scan(&count)
	return count, err
}

func (r *Repository) GetAdjustmentsByException(ctx context.Context, exceptionID string) ([]SettlementAdjustment, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, exception_case_id, amount, direction, reason, approved_by, journal_entry_id, created_at
		 FROM settlement_adjustments WHERE exception_case_id = $1 ORDER BY created_at`, exceptionID,
	)
	if err != nil {
		return nil, fmt.Errorf("get adjustments: %w", err)
	}
	defer rows.Close()

	var adjustments []SettlementAdjustment
	for rows.Next() {
		var sa SettlementAdjustment
		if err := rows.Scan(&sa.ID, &sa.ExceptionCaseID, &sa.Amount, &sa.Direction, &sa.Reason, &sa.ApprovedBy, &sa.JournalEntryID, &sa.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan adjustment: %w", err)
		}
		adjustments = append(adjustments, sa)
	}
	return adjustments, rows.Err()
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

func (r *Repository) GetQuotesBySupplier(ctx context.Context, supplierID string) ([]SupplierQuoteView, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT q.id, q.rfq_id, rfq.title, rfq.status, rfq.deadline, q.total_amount, q.submitted_at
		 FROM rfq_quotes q
		 JOIN rfqs rfq ON rfq.id = q.rfq_id
		 WHERE q.supplier_id = $1
		 ORDER BY q.submitted_at DESC`, supplierID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []SupplierQuoteView
	for rows.Next() {
		var v SupplierQuoteView
		var deadline time.Time
		var submittedAt *time.Time
		if err := rows.Scan(&v.ID, &v.RFQID, &v.RFQTitle, &v.RFQStatus, &deadline, &v.TotalAmount, &submittedAt); err != nil {
			return nil, err
		}
		v.RFQDeadline = deadline.Format(time.RFC3339)
		if submittedAt != nil {
			s := submittedAt.Format(time.RFC3339)
			v.SubmittedAt = &s
		}
		v.Status = "submitted"
		results = append(results, v)
	}
	if results == nil {
		results = []SupplierQuoteView{}
	}
	return results, rows.Err()
}

var _ = isDuplicateKey
