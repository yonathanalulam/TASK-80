DO $$ BEGIN
    CREATE TYPE rfq_status AS ENUM (
        'draft', 'issued', 'responded', 'comparison_ready',
        'selected', 'closed_no_award', 'converted_to_po'
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
    CREATE TYPE po_status AS ENUM (
        'draft', 'issued', 'accepted', 'partially_delivered',
        'delivered', 'inspection_pending', 'exception_open', 'closed'
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
    CREATE TYPE inspection_status AS ENUM ('pending', 'passed', 'failed');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
    CREATE TYPE discrepancy_type AS ENUM (
        'shortage', 'damage', 'wrong_item',
        'late_delivery', 'service_deviation', 'other'
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
    CREATE TYPE exception_status AS ENUM (
        'open', 'pending_financial_resolution',
        'pending_waiver', 'ready_to_close', 'closed'
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

CREATE TABLE IF NOT EXISTS rfqs (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_by  UUID NOT NULL REFERENCES users(id),
    title       VARCHAR NOT NULL,
    description TEXT,
    deadline    TIMESTAMPTZ,
    status      rfq_status NOT NULL DEFAULT 'draft',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_rfqs_created_by ON rfqs (created_by);
CREATE INDEX IF NOT EXISTS idx_rfqs_status ON rfqs (status);
CREATE INDEX IF NOT EXISTS idx_rfqs_deadline ON rfqs (deadline);

CREATE TABLE IF NOT EXISTS rfq_items (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rfq_id         UUID NOT NULL REFERENCES rfqs(id) ON DELETE CASCADE,
    item_name      VARCHAR NOT NULL,
    specifications TEXT,
    quantity       INT NOT NULL,
    unit           VARCHAR,
    sort_order     INT,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_rfq_items_rfq ON rfq_items (rfq_id);

CREATE TABLE IF NOT EXISTS rfq_suppliers (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rfq_id      UUID NOT NULL REFERENCES rfqs(id) ON DELETE CASCADE,
    supplier_id UUID NOT NULL REFERENCES users(id),
    invited_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_rfq_suppliers_rfq ON rfq_suppliers (rfq_id);
CREATE INDEX IF NOT EXISTS idx_rfq_suppliers_supplier ON rfq_suppliers (supplier_id);

CREATE TABLE IF NOT EXISTS rfq_quotes (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rfq_id        UUID NOT NULL REFERENCES rfqs(id),
    supplier_id   UUID NOT NULL REFERENCES users(id),
    total_amount  NUMERIC(18,2) NOT NULL,
    lead_time_days INT,
    notes         TEXT,
    submitted_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_rfq_quotes_rfq ON rfq_quotes (rfq_id);
CREATE INDEX IF NOT EXISTS idx_rfq_quotes_supplier ON rfq_quotes (supplier_id);

CREATE TABLE IF NOT EXISTS rfq_quote_items (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    quote_id    UUID NOT NULL REFERENCES rfq_quotes(id) ON DELETE CASCADE,
    rfq_item_id UUID NOT NULL REFERENCES rfq_items(id),
    unit_price  NUMERIC(18,2) NOT NULL,
    quantity    INT NOT NULL,
    subtotal    NUMERIC(18,2) NOT NULL,
    notes       TEXT
);

CREATE INDEX IF NOT EXISTS idx_rfq_quote_items_quote ON rfq_quote_items (quote_id);

CREATE TABLE IF NOT EXISTS purchase_orders (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rfq_id        UUID,
    quote_id      UUID,
    supplier_id   UUID NOT NULL REFERENCES users(id),
    created_by    UUID NOT NULL REFERENCES users(id),
    po_number     VARCHAR NOT NULL UNIQUE,
    promised_date TIMESTAMPTZ,
    status        po_status NOT NULL DEFAULT 'draft',
    total_amount  NUMERIC(18,2) NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_purchase_orders_supplier ON purchase_orders (supplier_id);
CREATE INDEX IF NOT EXISTS idx_purchase_orders_created_by ON purchase_orders (created_by);
CREATE INDEX IF NOT EXISTS idx_purchase_orders_status ON purchase_orders (status);

CREATE TABLE IF NOT EXISTS po_items (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    po_id          UUID NOT NULL REFERENCES purchase_orders(id) ON DELETE CASCADE,
    item_name      VARCHAR NOT NULL,
    specifications TEXT,
    unit_price     NUMERIC(18,2) NOT NULL,
    quantity       INT NOT NULL,
    subtotal       NUMERIC(18,2) NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_po_items_po ON po_items (po_id);

CREATE TABLE IF NOT EXISTS deliveries (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    po_id         UUID NOT NULL REFERENCES purchase_orders(id),
    courier_id    UUID,
    received_by   UUID,
    delivery_date TIMESTAMPTZ,
    notes         TEXT,
    status        VARCHAR NOT NULL DEFAULT 'pending',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_deliveries_po ON deliveries (po_id);
CREATE INDEX IF NOT EXISTS idx_deliveries_status ON deliveries (status);

CREATE TABLE IF NOT EXISTS delivery_items (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    delivery_id        UUID NOT NULL REFERENCES deliveries(id) ON DELETE CASCADE,
    po_item_id         UUID NOT NULL REFERENCES po_items(id),
    quantity_delivered  INT NOT NULL,
    quantity_accepted   INT NOT NULL DEFAULT 0,
    quantity_rejected   INT NOT NULL DEFAULT 0,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_delivery_items_delivery ON delivery_items (delivery_id);

CREATE TABLE IF NOT EXISTS quality_inspections (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    delivery_id  UUID REFERENCES deliveries(id),
    po_id        UUID NOT NULL REFERENCES purchase_orders(id),
    inspector_id UUID NOT NULL REFERENCES users(id),
    status       inspection_status NOT NULL DEFAULT 'pending',
    notes        TEXT,
    inspected_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_quality_inspections_po ON quality_inspections (po_id);
CREATE INDEX IF NOT EXISTS idx_quality_inspections_delivery ON quality_inspections (delivery_id);
CREATE INDEX IF NOT EXISTS idx_quality_inspections_status ON quality_inspections (status);

CREATE TABLE IF NOT EXISTS discrepancy_tickets (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    po_id            UUID NOT NULL REFERENCES purchase_orders(id),
    delivery_id      UUID,
    inspection_id    UUID,
    discrepancy_type discrepancy_type NOT NULL,
    description      TEXT NOT NULL,
    notes            TEXT,
    status           VARCHAR NOT NULL DEFAULT 'open',
    created_by       UUID NOT NULL REFERENCES users(id),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_discrepancy_tickets_po ON discrepancy_tickets (po_id);
CREATE INDEX IF NOT EXISTS idx_discrepancy_tickets_status ON discrepancy_tickets (status);

CREATE TABLE IF NOT EXISTS exception_cases (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reference_type VARCHAR NOT NULL,
    reference_id   UUID NOT NULL,
    status         exception_status NOT NULL DEFAULT 'open',
    opened_reason  TEXT,
    opened_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    closed_at      TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_exception_cases_reference
    ON exception_cases (reference_type, reference_id);
CREATE INDEX IF NOT EXISTS idx_exception_cases_status ON exception_cases (status);

CREATE TABLE IF NOT EXISTS waiver_records (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    exception_case_id UUID NOT NULL REFERENCES exception_cases(id),
    approved_by       UUID NOT NULL REFERENCES users(id),
    waiver_reason     TEXT,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_waiver_records_exception ON waiver_records (exception_case_id);

CREATE TABLE IF NOT EXISTS settlement_adjustments (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    exception_case_id UUID NOT NULL REFERENCES exception_cases(id),
    amount            NUMERIC(18,2) NOT NULL,
    direction         VARCHAR NOT NULL CHECK (direction IN ('debit', 'credit')),
    reason            TEXT,
    approved_by       UUID NOT NULL REFERENCES users(id),
    journal_entry_id  UUID,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_settlement_adjustments_exception
    ON settlement_adjustments (exception_case_id);
