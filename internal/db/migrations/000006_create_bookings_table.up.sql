CREATE TABLE bookings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
    slot_id UUID NOT NULL REFERENCES slots (id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    status VARCHAR(10) NOT NULL DEFAULT 'active',
    conference_link TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT chk_bookings_status CHECK (status IN ('active', 'cancelled'))
);

CREATE INDEX idx_bookings_slot_id ON bookings (slot_id);
