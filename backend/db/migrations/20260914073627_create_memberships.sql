-- +goose Up
CREATE TABLE memberships (
    id          SERIAL PRIMARY KEY,
    role_id     INTEGER NOT NULL REFERENCES socroles(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL, -- provided by ory
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),

    UNIQUE (role_id, user_id)
);

CREATE INDEX idx_memberships_user_id ON memberships (user_id);

-- +goose Down
DROP TABLE memberships;