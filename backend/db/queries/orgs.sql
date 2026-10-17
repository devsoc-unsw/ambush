-- name: CreateOrg :one
INSERT INTO orgs (name, slug)
VALUES ($1, $2)
RETURNING *;

-- name: GetOrg :one
SELECT * FROM orgs
WHERE id = $1;

-- TODO: GetOrgBySlug :one

-- TODO: ListOrgs :many

-- name: UpdateOrg :one
UPDATE orgs
SET name = COALESCE(sqlc.narg('name'), name),
    slug = COALESCE(sqlc.narg('slug'), slug),
    updated_at = now()
WHERE id = sqlc.arg('id') AND archived_at IS NULL
RETURNING *;

-- COALESCE keeps the original time, so archiving twice is a no-op
-- name: ArchiveOrg :one
UPDATE orgs
SET archived_at = COALESCE(archived_at, now())
WHERE id = $1
RETURNING *;
