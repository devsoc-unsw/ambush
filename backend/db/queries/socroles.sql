-- name: CreateSocRole :one
INSERT INTO socroles (soc_id, parent_role_id, name)
VALUES ($1, $2, $3)
RETURNING *;

-- TODO: GetSocRole :one

-- TODO: GetSocRoleInSociety :one

-- locks the role row so concurrent handovers of the same role run one at a
-- time. NO KEY UPDATE still lets memberships FK checks through
-- name: LockSocRoleInSociety :one
SELECT id FROM socroles
WHERE id = sqlc.arg('role_id') AND soc_id = sqlc.arg('soc_id')
FOR NO KEY UPDATE;

-- TODO: ListSocRolesBySociety :many

-- TODO: UpdateSocRole :one

-- TODO: DeleteSocRole :exec

-- TODO: ListRoleAncestors :many

-- TODO: ListRoleSubtree :many
