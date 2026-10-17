-- +goose Up
CREATE TABLE memberships (
    id          UUID PRIMARY KEY DEFAULT uuidv7(),
    -- RESTRICT so deleting a role can't wipe membership history. archive the
    -- role instead
    role_id     UUID NOT NULL REFERENCES roles(id) ON DELETE RESTRICT,
    user_id     UUID NOT NULL,  -- provided by ory
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    ended_at    TIMESTAMPTZ,    -- NULL = currently holds the role
    term        TEXT            -- label only, e.g. '2026' or '2026 T1'
);

-- only one active row per (role, user), so ended rows can stay as history and
-- the same person can hold the role again in a later term
CREATE UNIQUE INDEX memberships_active_role_user_uniq
    ON memberships (role_id, user_id) WHERE ended_at IS NULL;

CREATE INDEX idx_memberships_user_id ON memberships (user_id);

-- +goose Down
DROP TABLE memberships;
