-- +goose Up

CREATE TABLE roles (
    id          UUID PRIMARY KEY,
    org_id      UUID NOT NULL,
    name        TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    archived_at TIMESTAMPTZ
);

-- +goose Down
DROP TABLE roles;
