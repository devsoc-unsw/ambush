-- name: GetTeam :one
SELECT * FROM teams
WHERE id = $1;

-- name: CreateTeam :one
INSERT INTO teams (id, name)
VALUES ($1, $2)
RETURNING *;

-- name: DeleteTeam :exec
DELETE FROM teams
WHERE id = $1;
