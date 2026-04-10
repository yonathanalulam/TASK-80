DO $$ BEGIN
    CREATE TYPE wallet_type AS ENUM (
        'customer', 'supplier', 'courier',
        'platform_clearing', 'escrow_control',
        'refund_clearing', 'fee_revenue'
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
    CREATE TYPE escrow_status AS ENUM ('held', 'partially_released', 'released', 'refunded');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
    CREATE TYPE tender_type AS ENUM (
        'cash', 'card_on_file_recorded', 'bank_transfer_recorded', 'other_manual'
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
    CREATE TYPE withdrawal_status AS ENUM (
        'requested', 'under_review', 'approved', 'rejected', 'settled'
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

CREATE TABLE IF NOT EXISTS wallets (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id    UUID REFERENCES users(id),
    wallet_type wallet_type NOT NULL,
    balance     NUMERIC(18,2) NOT NULL DEFAULT 0,
    currency    VARCHAR(3) NOT NULL DEFAULT 'USD',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_wallets_owner ON wallets (owner_id);
CREATE INDEX IF NOT EXISTS idx_wallets_type ON wallets (wallet_type);

CREATE TABLE IF NOT EXISTS wallet_transactions (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id      UUID NOT NULL REFERENCES wallets(id),
    amount         NUMERIC(18,2) NOT NULL,
    direction      VARCHAR NOT NULL CHECK (direction IN ('credit', 'debit')),
    reference_type VARCHAR,
    reference_id   UUID,
    description    TEXT,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_wallet_transactions_wallet ON wallet_transactions (wallet_id, created_at);
CREATE INDEX IF NOT EXISTS idx_wallet_transactions_reference
    ON wallet_transactions (reference_type, reference_id);

CREATE TABLE IF NOT EXISTS escrow_accounts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_type      VARCHAR NOT NULL,
    order_id        UUID NOT NULL,
    amount_held     NUMERIC(18,2) NOT NULL DEFAULT 0,
    amount_released NUMERIC(18,2) NOT NULL DEFAULT 0,
    amount_refunded NUMERIC(18,2) NOT NULL DEFAULT 0,
    status          escrow_status NOT NULL DEFAULT 'held',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_escrow_accounts_order
    ON escrow_accounts (order_type, order_id);
CREATE INDEX IF NOT EXISTS idx_escrow_accounts_status ON escrow_accounts (status);

CREATE TABLE IF NOT EXISTS payment_records (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_type     VARCHAR NOT NULL,
    order_id       UUID NOT NULL,
    tender_type    tender_type NOT NULL,
    amount         NUMERIC(18,2) NOT NULL,
    currency       VARCHAR(3) NOT NULL DEFAULT 'USD',
    reference_text TEXT,
    recorded_by    UUID NOT NULL REFERENCES users(id),
    recorded_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payment_records_order
    ON payment_records (order_type, order_id);

CREATE TABLE IF NOT EXISTS refunds (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_type    VARCHAR NOT NULL,
    order_id      UUID NOT NULL,
    refund_amount NUMERIC(18,2) NOT NULL,
    refund_reason TEXT,
    created_by    UUID NOT NULL REFERENCES users(id),
    approved_by   UUID,
    status        VARCHAR NOT NULL DEFAULT 'pending',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_refunds_order ON refunds (order_type, order_id);
CREATE INDEX IF NOT EXISTS idx_refunds_status ON refunds (status);

CREATE TABLE IF NOT EXISTS refund_items (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    refund_id  UUID NOT NULL REFERENCES refunds(id) ON DELETE CASCADE,
    item_id    UUID NOT NULL,
    item_type  VARCHAR NOT NULL,
    amount     NUMERIC(18,2) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_refund_items_refund ON refund_items (refund_id);

CREATE TABLE IF NOT EXISTS withdrawal_requests (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    courier_id      UUID NOT NULL REFERENCES users(id),
    request_amount  NUMERIC(18,2) NOT NULL,
    status          withdrawal_status NOT NULL DEFAULT 'requested',
    requested_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    reviewed_by     UUID,
    approved_by     UUID,
    rejected_reason TEXT,
    settled_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_withdrawal_requests_courier ON withdrawal_requests (courier_id);
CREATE INDEX IF NOT EXISTS idx_withdrawal_requests_status ON withdrawal_requests (status);

CREATE TABLE IF NOT EXISTS withdrawal_disbursements (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    withdrawal_id UUID NOT NULL REFERENCES withdrawal_requests(id),
    amount        NUMERIC(18,2) NOT NULL,
    disbursed_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_withdrawal_disbursements_withdrawal
    ON withdrawal_disbursements (withdrawal_id);

CREATE TABLE IF NOT EXISTS journal_entries (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entry_type     VARCHAR NOT NULL,
    reference_type VARCHAR,
    reference_id   UUID,
    description    TEXT,
    effective_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by     UUID,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_journal_entries_reference
    ON journal_entries (reference_type, reference_id);
CREATE INDEX IF NOT EXISTS idx_journal_entries_effective ON journal_entries (effective_at);

CREATE TABLE IF NOT EXISTS journal_lines (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    journal_entry_id UUID NOT NULL REFERENCES journal_entries(id),
    account_code     VARCHAR NOT NULL,
    direction        VARCHAR NOT NULL CHECK (direction IN ('debit', 'credit')),
    amount           NUMERIC(18,2) NOT NULL,
    counterparty_id  UUID,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_journal_lines_entry ON journal_lines (journal_entry_id);
CREATE INDEX IF NOT EXISTS idx_journal_lines_account ON journal_lines (account_code);

CREATE TABLE IF NOT EXISTS counterparties (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       VARCHAR NOT NULL,
    type       VARCHAR NOT NULL,
    user_id    UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_counterparties_user ON counterparties (user_id);

CREATE TABLE IF NOT EXISTS reconciliation_runs (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    run_date     DATE NOT NULL,
    status       VARCHAR NOT NULL DEFAULT 'pending',
    summary_json JSONB,
    created_by   UUID,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_reconciliation_runs_date ON reconciliation_runs (run_date);

CREATE TABLE IF NOT EXISTS reconciliation_items (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    run_id          UUID NOT NULL REFERENCES reconciliation_runs(id) ON DELETE CASCADE,
    item_type       VARCHAR NOT NULL,
    reference_id    UUID NOT NULL,
    expected_amount NUMERIC(18,2) NOT NULL,
    actual_amount   NUMERIC(18,2) NOT NULL,
    difference      NUMERIC(18,2) NOT NULL,
    status          VARCHAR NOT NULL,
    notes           TEXT
);

CREATE INDEX IF NOT EXISTS idx_reconciliation_items_run ON reconciliation_items (run_id);
