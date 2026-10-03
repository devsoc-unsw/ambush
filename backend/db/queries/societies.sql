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
WHERE id = sqlc.arg('id') AND archived_at IS NULL
RETURNING *;

-- COALESCE keeps the original time, so archiving twice is a no-op
-- name: ArchiveSociety :one
UPDATE societies
SET archived_at = COALESCE(archived_at, now())
WHERE id = $1
RETURNING *;