-- +goose Up
CREATE TABLE socroles (
    id              SERIAL PRIMARY KEY,
    soc_id          INTEGER NOT NULL REFERENCES societies(id) ON DELETE CASCADE,
    parent_role_id  INTEGER,
    name            TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),

    UNIQUE (soc_id, name),          -- role names unique per soc
    UNIQUE (soc_id, id),            -- role from soc A cannot be parent of role from soc B
    CHECK (parent_role_id <> id),   -- ensures that parent is not itself

    FOREIGN KEY (soc_id, parent_role_id)
        REFERENCES socroles (soc_id, id)
        ON DELETE RESTRICT          -- make sure a deletion of parent doesnt delete the whole row
);

CREATE INDEX idx_socroles_soc_id ON socroles (soc_id);
CREATE INDEX idx_socroles_parent ON socroles (parent_role_id);

-- +goose Down
DROP TABLE socroles;