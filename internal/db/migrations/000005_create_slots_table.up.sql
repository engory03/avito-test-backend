CREATE TABLE slots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
    room_id UUID NOT NULL REFERENCES rooms (id) ON DELETE CASCADE,
    start_at TIMESTAMPTZ NOT NULL,
    end_at TIMESTAMPTZ NOT NULL,
    CONSTRAINT chk_slots_end CHECK (end_at > start_at)
);

CREATE UNIQUE INDEX idx_room_start ON slots (room_id, start_at);
