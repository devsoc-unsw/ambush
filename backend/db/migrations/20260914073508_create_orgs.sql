-- +goose Up
CREATE TABLE orgs (
    id          UUID PRIMARY KEY DEFAULT uuidv7(),
    name        TEXT NOT NULL,
    slug        TEXT NOT NULL UNIQUE,  -- stays reserved after archiving
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    archived_at TIMESTAMPTZ            -- NULL = active. archived orgs are read only
);

-- +goose Down
DROP TABLE orgs;
