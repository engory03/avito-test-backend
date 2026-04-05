CREATE TABLE schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
    room_id UUID NOT NULL REFERENCES rooms (id) ON DELETE CASCADE,
    days_of_week INTEGER[] NOT NULL,
    start_time VARCHAR(5) NOT NULL,
    end_time VARCHAR(5) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_schedules_room_id ON schedules (room_id);
