DO $$ BEGIN
    CREATE TYPE booking_status AS ENUM (
        'draft', 'pending_pricing', 'pending_payment_record',
        'paid_held_in_escrow', 'confirmed', 'partially_fulfilled',
        'fulfilled', 'completed', 'cancelled',
        'refunded_partial', 'refunded_full', 'closed'
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

CREATE TABLE IF NOT EXISTS bookings (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organizer_id        UUID NOT NULL REFERENCES users(id),
    itinerary_id        UUID REFERENCES itineraries(id),
    title               VARCHAR NOT NULL,
    description         TEXT,
    status              booking_status NOT NULL DEFAULT 'draft',
    total_amount        NUMERIC(18,2),
    discount_amount     NUMERIC(18,2) NOT NULL DEFAULT 0,
    escrow_amount       NUMERIC(18,2) NOT NULL DEFAULT 0,
    pricing_snapshot_id UUID,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_bookings_organizer ON bookings (organizer_id);
CREATE INDEX IF NOT EXISTS idx_bookings_itinerary ON bookings (itinerary_id);
CREATE INDEX IF NOT EXISTS idx_bookings_status ON bookings (status);
CREATE INDEX IF NOT EXISTS idx_bookings_created_at ON bookings (created_at);

CREATE TABLE IF NOT EXISTS booking_items (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    booking_id  UUID NOT NULL REFERENCES bookings(id) ON DELETE CASCADE,
    item_type   VARCHAR NOT NULL,
    item_name   VARCHAR NOT NULL,
    description TEXT,
    unit_price  NUMERIC(18,2) NOT NULL,
    quantity    INT NOT NULL DEFAULT 1,
    subtotal    NUMERIC(18,2) NOT NULL,
    category    VARCHAR,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_booking_items_booking ON booking_items (booking_id);

CREATE TABLE IF NOT EXISTS checkout_pricing_snapshots (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    booking_id           UUID REFERENCES bookings(id),
    procurement_order_id UUID,
    snapshot_json        JSONB NOT NULL,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_checkout_pricing_snapshots_booking
    ON checkout_pricing_snapshots (booking_id);
CREATE INDEX IF NOT EXISTS idx_checkout_pricing_snapshots_po
    ON checkout_pricing_snapshots (procurement_order_id);
