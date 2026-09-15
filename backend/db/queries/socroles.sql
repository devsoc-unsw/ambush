-- name: CreateSocRole :one
INSERT INTO socroles (soc_id, parent_role_id, name)
VALUES ($1, $2, $3)
RETURNING *;

-- TODO: GetSocRole :one

-- TODO: GetSocRoleInSociety :one

-- TODO: ListSocRolesBySociety :many

-- TODO: UpdateSocRole :one

-- TODO: DeleteSocRole :exec

-- TODO: ListRoleAncestors :many

-- TODO: ListRoleSubtree :many
