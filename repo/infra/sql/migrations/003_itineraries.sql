DO $$ BEGIN
    CREATE TYPE itinerary_status AS ENUM (
        'draft', 'published', 'revised', 'in_progress',
        'completed', 'cancelled', 'archived'
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

CREATE TABLE IF NOT EXISTS itineraries (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organizer_id        UUID NOT NULL REFERENCES users(id),
    title               VARCHAR NOT NULL,
    meetup_at           TIMESTAMPTZ,
    meetup_location_text VARCHAR,
    notes               TEXT,
    status              itinerary_status NOT NULL DEFAULT 'draft',
    published_at        TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_itineraries_organizer ON itineraries (organizer_id);
CREATE INDEX IF NOT EXISTS idx_itineraries_status ON itineraries (status);
CREATE INDEX IF NOT EXISTS idx_itineraries_meetup_at ON itineraries (meetup_at);

CREATE TABLE IF NOT EXISTS itinerary_checkpoints (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    itinerary_id  UUID NOT NULL REFERENCES itineraries(id) ON DELETE CASCADE,
    sort_order    INT NOT NULL,
    checkpoint_text VARCHAR NOT NULL,
    eta           TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_itinerary_checkpoints_itinerary
    ON itinerary_checkpoints (itinerary_id, sort_order);

CREATE TABLE IF NOT EXISTS itinerary_members (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    itinerary_id UUID NOT NULL REFERENCES itineraries(id) ON DELETE CASCADE,
    user_id      UUID NOT NULL REFERENCES users(id),
    role         VARCHAR NOT NULL DEFAULT 'participant',
    joined_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (itinerary_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_itinerary_members_user ON itinerary_members (user_id);

CREATE TABLE IF NOT EXISTS itinerary_member_form_definitions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    itinerary_id    UUID NOT NULL REFERENCES itineraries(id) ON DELETE CASCADE,
    field_key       VARCHAR NOT NULL,
    field_label     VARCHAR NOT NULL,
    field_type      VARCHAR NOT NULL,
    required        BOOLEAN NOT NULL DEFAULT false,
    options_json    JSONB,
    validation_json JSONB,
    active          BOOLEAN NOT NULL DEFAULT true,
    sort_order      INT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_itinerary_form_defs_itinerary
    ON itinerary_member_form_definitions (itinerary_id);

CREATE TABLE IF NOT EXISTS itinerary_member_form_submissions (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    itinerary_id   UUID NOT NULL REFERENCES itineraries(id) ON DELETE CASCADE,
    member_user_id UUID NOT NULL REFERENCES users(id),
    payload_json   JSONB,
    submitted_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_itinerary_form_subs_itinerary
    ON itinerary_member_form_submissions (itinerary_id);
CREATE INDEX IF NOT EXISTS idx_itinerary_form_subs_member
    ON itinerary_member_form_submissions (member_user_id);

CREATE TABLE IF NOT EXISTS itinerary_change_events (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    itinerary_id  UUID NOT NULL REFERENCES itineraries(id) ON DELETE CASCADE,
    actor_id      UUID NOT NULL REFERENCES users(id),
    change_type   VARCHAR NOT NULL,
    summary       TEXT,
    diff_json     JSONB,
    visible_from  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_itinerary_change_events_itinerary
    ON itinerary_change_events (itinerary_id, created_at);

CREATE TABLE IF NOT EXISTS itinerary_notifications (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    itinerary_id      UUID NOT NULL REFERENCES itineraries(id) ON DELETE CASCADE,
    recipient_user_id UUID NOT NULL REFERENCES users(id),
    message           TEXT NOT NULL,
    read_at           TIMESTAMPTZ,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_itinerary_notifications_recipient
    ON itinerary_notifications (recipient_user_id, read_at);
CREATE INDEX IF NOT EXISTS idx_itinerary_notifications_itinerary
    ON itinerary_notifications (itinerary_id);
