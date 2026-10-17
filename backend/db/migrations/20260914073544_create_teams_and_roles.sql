-- +goose Up
-- org -> teams (a tree) -> roles -> memberships. people hold roles, and are in
-- a team because they hold one of its roles
CREATE TABLE teams (
    id              UUID PRIMARY KEY DEFAULT uuidv7(),
    org_id          UUID NOT NULL REFERENCES orgs(id),
    parent_id       UUID,               -- NULL = the org's root team (exec)
    leader_role_id  UUID,               -- which role leads this team, FK below
    name            TEXT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    archived_at     TIMESTAMPTZ,

    UNIQUE (org_id, name),
    UNIQUE (org_id, id),                -- target for the same-org FKs
    CHECK (parent_id <> id),

    -- a parent from another org is rejected by the composite key
    FOREIGN KEY (org_id, parent_id) REFERENCES teams (org_id, id)
);

-- one root team per org, so "org head" = leader of the root team
CREATE UNIQUE INDEX teams_one_root_per_org ON teams (org_id)
    WHERE parent_id IS NULL;
CREATE INDEX idx_teams_parent ON teams (parent_id);

CREATE TABLE roles (
    id          UUID PRIMARY KEY DEFAULT uuidv7(),
    -- copied from the team so "is this role in this org" is one join, and so
    -- teams.leader_role_id can be checked against the same org. the FK below
    -- keeps it in sync with the team's real org
    org_id      UUID NOT NULL,
    team_id     UUID NOT NULL,
    name        TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    archived_at TIMESTAMPTZ,            -- retire roles instead of deleting them

    UNIQUE (team_id, name),
    UNIQUE (org_id, id),

    FOREIGN KEY (org_id, team_id) REFERENCES teams (org_id, id)
);

CREATE INDEX idx_roles_team ON roles (team_id);

-- teams and roles point at each other, so this FK is added once both exist.
-- the leader role can be from another team (VP Ops leading Platforms) but only
-- from the same org. deleting the role clears the slot rather than the team
ALTER TABLE teams ADD CONSTRAINT teams_leader_role_fkey
    FOREIGN KEY (org_id, leader_role_id) REFERENCES roles (org_id, id)
    ON DELETE SET NULL (leader_role_id);

-- +goose Down
-- one statement, since each table references the other
DROP TABLE roles, teams;
