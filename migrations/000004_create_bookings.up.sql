CREATE TABLE bookings (
    id UUID PRIMARY KEY,
    slot_id UUID NOT NULL REFERENCES slots(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    status TEXT NOT NULL,
    conference_link TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (status IN ('active', 'cancelled'))
);

CREATE UNIQUE INDEX uniq_active_booking_per_slot
    ON bookings (slot_id)
    WHERE status = 'active';