CREATE EXTENSION IF NOT EXISTS btree_gist;

CREATE TABLE slots (
    id UUID PRIMARY KEY,
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    start_at TIMESTAMPTZ NOT NULL,
    end_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (start_at < end_at),
    EXCLUDE USING gist (
        room_id WITH =,
        tstzrange(start_at, end_at, '[)') WITH &&
    )
);

CREATE INDEX idx_slots_room_id_start_at ON slots (room_id, start_at);