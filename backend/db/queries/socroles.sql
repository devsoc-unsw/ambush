-- selects from societies so nothing is inserted for a missing or archived one
-- name: CreateSocRole :one
INSERT INTO socroles (soc_id, parent_role_id, name)
SELECT s.id, sqlc.narg('parent_role_id')::int, sqlc.arg('name')::text
FROM societies s
WHERE s.id = sqlc.arg('soc_id') AND s.archived_at IS NULL
RETURNING *;

-- TODO: GetSocRole :one

-- TODO: GetSocRoleInSociety :one

-- locks the role row so concurrent handovers of the same role run one at a
-- time. NO KEY UPDATE still lets memberships FK checks through
-- name: LockSocRoleInSociety :one
SELECT r.id FROM socroles r
JOIN societies s ON s.id = r.soc_id AND s.archived_at IS NULL
WHERE r.id = sqlc.arg('role_id') AND r.soc_id = sqlc.arg('soc_id')
FOR NO KEY UPDATE OF r;

-- TODO: ListSocRolesBySociety :many

-- TODO: UpdateSocRole :one

-- TODO: DeleteSocRole :exec

-- TODO: ListRoleAncestors :many

-- TODO: ListRoleSubtree :many
