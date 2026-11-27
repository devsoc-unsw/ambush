-- name: GetRole :one
SELECT * FROM roles
WHERE id = $1;

-- name: CreateRole :one
INSERT INTO roles (id, name)
VALUES ($1, $2)
RETURNING *;

-- name: DeleteRole :exec
DELETE FROM roles
WHERE id = $1;
