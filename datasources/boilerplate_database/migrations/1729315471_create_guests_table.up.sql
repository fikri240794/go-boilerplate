CREATE TABLE guests (
    id uuid primary key,
    name text not null,
    address text,
    created_at bigint not null,
    created_by text not null,
    updated_at bigint,
    updated_by text,
    deleted_at bigint,
    deleted_by text
);