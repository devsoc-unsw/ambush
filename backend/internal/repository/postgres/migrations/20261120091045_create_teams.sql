-- +goose Up

CREATE TABLE teams (
    id              UUID PRIMARY KEY DEFAULT uuidv7(),
    org_id          UUID NOT NULL REFERENCES orgs(id),
    name            TEXT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    archived_at     TIMESTAMPTZ,

    UNIQUE (org_id, name),
    UNIQUE (org_id, id)         -- target for the same-org FKs
);


-- +goose Down
DROP TABLE teams;
