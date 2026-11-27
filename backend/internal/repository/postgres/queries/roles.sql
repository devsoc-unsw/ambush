-- name: GetRole :one
SELECT * FROM roles
WHERE id = $1;
