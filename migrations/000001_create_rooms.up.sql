create table if not exists rooms (
    id uuid primary key,
    name text not null,
    description text,
    capacity integer not null check (capacity > 0),
    created_at timestamptz not null default now()
);