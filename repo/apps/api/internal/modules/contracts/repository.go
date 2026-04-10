package contracts

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
func (r *Repository) GetActiveTemplates(ctx context.Context) ([]ContractTemplate, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, body_template, variable_schema_json, active, version, created_at, updated_at
		 FROM contract_templates WHERE active = TRUE ORDER BY name`,
	)
	if err != nil {
		return nil, fmt.Errorf("list templates: %w", err)
	}
	defer rows.Close()

	var items []ContractTemplate
	for rows.Next() {
		var t ContractTemplate
		if err := rows.Scan(
			&t.ID, &t.Name, &t.BodyTemplate, &t.VariableSchemaJSON,
			&t.Active, &t.Version, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan template: %w", err)
		}
		items = append(items, t)
	}
	return items, rows.Err()
}

func (r *Repository) GetTemplateByID(ctx context.Context, id string) (*ContractTemplate, error) {
	t := &ContractTemplate{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, body_template, variable_schema_json, active, version, created_at, updated_at
		 FROM contract_templates WHERE id = $1`, id,
	).Scan(
		&t.ID, &t.Name, &t.BodyTemplate, &t.VariableSchemaJSON,
		&t.Active, &t.Version, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, common.NewNotFoundError("contract template")
		}
		return nil, fmt.Errorf("get template: %w", err)
	}
	return t, nil
}
func (r *Repository) CreateGeneratedContract(ctx context.Context, gc *GeneratedContract) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO generated_contracts (id, template_id, variables_json, file_id, generated_by, generated_at, version)
		 VALUES (gen_random_uuid(), $1, $2, $3, $4, NOW(), $5)
		 RETURNING id, generated_at`,
		gc.TemplateID, gc.VariablesJSON, gc.FileID, gc.GeneratedBy, gc.Version,
	).Scan(&gc.ID, &gc.GeneratedAt)
}
func (r *Repository) CreateInvoiceRequest(ctx context.Context, ir *InvoiceRequest) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO invoice_requests (id, requester_id, order_type, order_id, status, notes, created_at, updated_at)
		 VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, NOW(), NOW())
		 RETURNING id, created_at, updated_at`,
		ir.RequesterID, ir.OrderType, ir.OrderID, ir.Status, ir.Notes,
	).Scan(&ir.ID, &ir.CreatedAt, &ir.UpdatedAt)
}

func (r *Repository) GetInvoiceRequestByID(ctx context.Context, id string) (*InvoiceRequest, error) {
	ir := &InvoiceRequest{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, requester_id, order_type, order_id, status, notes, created_at, updated_at
		 FROM invoice_requests WHERE id = $1`, id,
	).Scan(&ir.ID, &ir.RequesterID, &ir.OrderType, &ir.OrderID, &ir.Status, &ir.Notes, &ir.CreatedAt, &ir.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, common.NewNotFoundError("invoice request")
		}
		return nil, fmt.Errorf("get invoice request: %w", err)
	}
	return ir, nil
}

func (r *Repository) UpdateInvoiceRequestStatus(ctx context.Context, id string, status InvoiceRequestStatus) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE invoice_requests SET status = $2, updated_at = NOW() WHERE id = $1`,
		id, status,
	)
	if err != nil {
		return fmt.Errorf("update invoice request status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return common.NewNotFoundError("invoice request")
	}
	return nil
}

func (r *Repository) ListInvoiceRequests(ctx context.Context, requesterID string, isAdmin bool) ([]InvoiceRequest, error) {
	var rows pgx.Rows
	var err error

	if isAdmin {
		rows, err = r.pool.Query(ctx,
			`SELECT id, requester_id, order_type, order_id, status, notes, created_at, updated_at
			 FROM invoice_requests ORDER BY created_at DESC`,
		)
	} else {
		rows, err = r.pool.Query(ctx,
			`SELECT id, requester_id, order_type, order_id, status, notes, created_at, updated_at
			 FROM invoice_requests WHERE requester_id = $1 ORDER BY created_at DESC`,
			requesterID,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("list invoice requests: %w", err)
	}
	defer rows.Close()

	var items []InvoiceRequest
	for rows.Next() {
		var ir InvoiceRequest
		if err := rows.Scan(&ir.ID, &ir.RequesterID, &ir.OrderType, &ir.OrderID, &ir.Status, &ir.Notes, &ir.CreatedAt, &ir.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan invoice request: %w", err)
		}
		items = append(items, ir)
	}
	return items, rows.Err()
}
func (r *Repository) CreateInvoice(ctx context.Context, inv *Invoice) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO invoices (id, request_id, invoice_number, order_type, order_id, amount, file_id, generated_at, created_at)
		 VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, NOW(), NOW())
		 RETURNING id, generated_at, created_at`,
		inv.RequestID, inv.InvoiceNumber, inv.OrderType, inv.OrderID, inv.Amount, inv.FileID,
	).Scan(&inv.ID, &inv.GeneratedAt, &inv.CreatedAt)
}

func (r *Repository) GetNextInvoiceSequence(ctx context.Context, year int) (int, error) {
	var count int
	prefix := fmt.Sprintf("INV-%d-%%", year)
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM invoices WHERE invoice_number LIKE $1`, prefix,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count invoices for year: %w", err)
	}
	return count + 1, nil
}

var _ = time.Now
