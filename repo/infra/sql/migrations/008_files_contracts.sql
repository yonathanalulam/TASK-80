DO $$ BEGIN
    CREATE TYPE invoice_request_status AS ENUM (
        'requested', 'approved', 'generated', 'delivered', 'cancelled'
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

CREATE TABLE IF NOT EXISTS file_metadata (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    storage_key          VARCHAR NOT NULL UNIQUE,
    original_filename    VARCHAR,
    mime_type            VARCHAR,
    byte_size            BIGINT,
    sha256               VARCHAR,
    encrypted            BOOLEAN NOT NULL DEFAULT false,
    encryption_key_wrapped TEXT,
    owner_user_id        UUID NOT NULL REFERENCES users(id),
    visibility_scope     VARCHAR NOT NULL DEFAULT 'private',
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_file_metadata_owner ON file_metadata (owner_user_id);
CREATE INDEX IF NOT EXISTS idx_file_metadata_visibility ON file_metadata (visibility_scope);

CREATE TABLE IF NOT EXISTS file_record_links (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    file_id     UUID NOT NULL REFERENCES file_metadata(id) ON DELETE CASCADE,
    record_type VARCHAR NOT NULL,
    record_id   UUID NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_file_record_links_file ON file_record_links (file_id);
CREATE INDEX IF NOT EXISTS idx_file_record_links_record
    ON file_record_links (record_type, record_id);

CREATE TABLE IF NOT EXISTS file_access_policies (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    file_id    UUID NOT NULL REFERENCES file_metadata(id) ON DELETE CASCADE,
    role       VARCHAR NOT NULL,
    permission VARCHAR NOT NULL DEFAULT 'read',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_file_access_policies_file ON file_access_policies (file_id);

CREATE TABLE IF NOT EXISTS download_tokens (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    token       VARCHAR NOT NULL UNIQUE,
    file_id     UUID NOT NULL REFERENCES file_metadata(id),
    actor_id    UUID NOT NULL REFERENCES users(id),
    expires_at  TIMESTAMPTZ NOT NULL,
    consumed_at TIMESTAMPTZ,
    single_use  BOOLEAN NOT NULL DEFAULT true,
    scope       VARCHAR,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_download_tokens_file ON download_tokens (file_id);
CREATE INDEX IF NOT EXISTS idx_download_tokens_expires ON download_tokens (expires_at);
CREATE INDEX IF NOT EXISTS idx_download_tokens_token ON download_tokens (token);

CREATE TABLE IF NOT EXISTS contract_templates (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name                 VARCHAR NOT NULL,
    body_template        TEXT,
    variable_schema_json JSONB,
    active               BOOLEAN NOT NULL DEFAULT true,
    version              INT NOT NULL DEFAULT 1,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS generated_contracts (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_id    UUID NOT NULL REFERENCES contract_templates(id),
    variables_json JSONB,
    file_id        UUID REFERENCES file_metadata(id),
    generated_by   UUID NOT NULL REFERENCES users(id),
    generated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    version        INT NOT NULL DEFAULT 1
);

CREATE INDEX IF NOT EXISTS idx_generated_contracts_template ON generated_contracts (template_id);

CREATE TABLE IF NOT EXISTS invoice_requests (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    requester_id UUID NOT NULL REFERENCES users(id),
    order_type   VARCHAR NOT NULL,
    order_id     UUID NOT NULL,
    status       invoice_request_status NOT NULL DEFAULT 'requested',
    notes        TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_invoice_requests_requester ON invoice_requests (requester_id);
CREATE INDEX IF NOT EXISTS idx_invoice_requests_order
    ON invoice_requests (order_type, order_id);
CREATE INDEX IF NOT EXISTS idx_invoice_requests_status ON invoice_requests (status);

CREATE TABLE IF NOT EXISTS invoices (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id     UUID REFERENCES invoice_requests(id),
    invoice_number VARCHAR NOT NULL UNIQUE,
    order_type     VARCHAR NOT NULL,
    order_id       UUID NOT NULL,
    amount         NUMERIC(18,2) NOT NULL,
    file_id        UUID REFERENCES file_metadata(id),
    generated_at   TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_invoices_order ON invoices (order_type, order_id);
CREATE INDEX IF NOT EXISTS idx_invoices_request ON invoices (request_id);
