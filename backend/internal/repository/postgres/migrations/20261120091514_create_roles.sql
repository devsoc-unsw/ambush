-- +goose Up

CREATE TABLE roles (
    id          UUID PRIMARY KEY DEFAULT uuidv7(),
    org_id      UUID NOT NULL,
    team_id     UUID NOT NULL,
    name        TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    archived_at TIMESTAMPTZ,

    UNIQUE (team_id, name),
    UNIQUE (org_id, id),

    FOREIGN KEY (org_id, team_id) REFERENCES teams (org_id, id)
);

-- +goose Down
DROP TABLE roles;
