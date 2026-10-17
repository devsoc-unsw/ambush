-- only used when creating an org. the unique index allows one root per org
-- name: CreateRootTeam :one
INSERT INTO teams (org_id, name)
VALUES ($1, $2)
RETURNING *;

-- selects the parent so nothing is inserted under a missing or archived team,
-- or in an archived org. the parent decides the org, so it can't be crossed
-- name: CreateTeam :one
INSERT INTO teams (org_id, parent_id, name)
SELECT p.org_id, p.id, sqlc.arg('name')::text
FROM teams p
JOIN orgs o ON o.id = p.org_id AND o.archived_at IS NULL
WHERE p.id = sqlc.arg('parent_id')
  AND p.org_id = sqlc.arg('org_id')
  AND p.archived_at IS NULL
RETURNING *;

-- name: GetTeam :one
SELECT * FROM teams
WHERE id = $1;

-- TODO: ListTeamsByOrg :many

-- TODO: ListTeamSubtree :many

-- sets or clears (NULL) the role that leads a team. the role can be from any
-- team in the same org, but not archived
-- name: SetTeamLeader :one
UPDATE teams t
SET leader_role_id = sqlc.narg('leader_role_id'),
    updated_at = now()
FROM orgs o
WHERE t.id = sqlc.arg('team_id')
  AND t.org_id = sqlc.arg('org_id')
  AND t.archived_at IS NULL
  AND o.id = t.org_id
  AND o.archived_at IS NULL
  AND (sqlc.narg('leader_role_id')::uuid IS NULL OR EXISTS (
      SELECT 1 FROM roles r
      WHERE r.id = sqlc.narg('leader_role_id')
        AND r.org_id = t.org_id
        AND r.archived_at IS NULL
  ))
RETURNING t.*;
