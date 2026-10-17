-- selects the team so nothing is inserted for a missing or archived team, or
-- in an archived org. org_id is copied from the team
-- name: CreateRole :one
INSERT INTO roles (org_id, team_id, name)
SELECT t.org_id, t.id, sqlc.arg('name')::text
FROM teams t
JOIN orgs o ON o.id = t.org_id AND o.archived_at IS NULL
WHERE t.id = sqlc.arg('team_id')
  AND t.org_id = sqlc.arg('org_id')
  AND t.archived_at IS NULL
RETURNING *;

-- TODO: GetRole :one

-- TODO: ListRolesByTeam :many

-- TODO: UpdateRole :one

-- TODO: ArchiveRole :one

-- locks the role row so concurrent handovers of the same role run one at a
-- time. NO KEY UPDATE still lets memberships FK checks through
-- name: LockActiveRoleInOrg :one
SELECT r.id FROM roles r
JOIN teams t ON t.id = r.team_id AND t.archived_at IS NULL
JOIN orgs o ON o.id = r.org_id AND o.archived_at IS NULL
WHERE r.id = sqlc.arg('role_id')
  AND r.org_id = sqlc.arg('org_id')
  AND r.archived_at IS NULL
FOR NO KEY UPDATE OF r;
