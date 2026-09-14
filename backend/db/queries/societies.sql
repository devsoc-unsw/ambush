-- name: CreateSocity :one
INSERT INTO societies (name, slug)
VALUES ($1, $2)
RETURNING *;

-- name: GetSociety :one

-- name: GetSocietyBySlug :one

-- name: ListSocieties :many

-- name: UpdateSociety :one
UPDATE societies
SET name = COALESCE(sqlc.narg('name'), name),
    slug = COALESCE(sqlc.narg('slug'), slug),
    updated_at = now()
WHERE id = sqlc.arg('id')
RETURNING *;