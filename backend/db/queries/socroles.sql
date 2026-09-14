-- name: CreateSocRole :one
INSERT INTO socroles (soc_id, parent_role_id, name)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetSocRole :one

-- name: GetSocRoleInSociety :one

-- name: ListSocRolesBySociety :many

-- name: UpdateSocRole :one

-- name: DeleteSocRole :exec

-- name: ListRoleAncestors :many

-- name: ListRoleSubtree :many
