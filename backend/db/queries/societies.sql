-- name: CreateSociety :one
INSERT INTO societies (name, slug)
VALUES ($1, $2)
RETURNING *;

-- name: GetSociety :one
SELECT * FROM societies
WHERE id = $1;

-- TODO: GetSocietyBySlug :one

-- TODO: ListSocieties :many

-- name: UpdateSociety :one
UPDATE societies
SET name = COALESCE(sqlc.narg('name'), name),
    slug = COALESCE(sqlc.narg('slug'), slug),
    updated_at = now()
WHERE id = sqlc.arg('id')
RETURNING *;