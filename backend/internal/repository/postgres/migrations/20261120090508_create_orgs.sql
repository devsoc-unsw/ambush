-- +goose Up

CREATE TABLE orgs (
    id          UUID PRIMARY KEY DEFAULT uuidv7(),
    name        TEXT NOT NULL,
    slug        TEXT NOT NULL UNIQUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    archived_at TIMESTAMPTZ
);

-- +goose Down
DROP TABLE orgs;
