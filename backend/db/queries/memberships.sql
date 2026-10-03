-- note: ensures the role does indeed belong to the society the user is being added to
-- name: CreateMembership :one
INSERT INTO memberships (role_id, user_id, term)
SELECT r.id, sqlc.arg('user_id'), sqlc.narg('term')::text
FROM socroles r
JOIN societies s ON s.id = r.soc_id AND s.archived_at IS NULL
WHERE r.id = sqlc.arg('role_id') AND r.soc_id = sqlc.arg('soc_id')
RETURNING *;

-- TODO: GetMembership :one

-- name: CreateMemberships :many
INSERT INTO memberships (role_id, user_id)
SELECT r.id, u.user_id
FROM (
    SELECT unnest(sqlc.arg('role_ids')::int[])  AS role_id,
           unnest(sqlc.arg('user_ids')::uuid[]) AS user_id
) AS u
JOIN socroles r ON r.id = u.role_id AND r.soc_id = sqlc.arg('soc_id')
JOIN societies s ON s.id = r.soc_id AND s.archived_at IS NULL
-- the WHERE has to match the partial unique index for postgres to find it
ON CONFLICT (role_id, user_id) WHERE ended_at IS NULL DO NOTHING
RETURNING *;

-- TODO: ListMembershipsBySociety :many

-- TODO: ListMembershipsByRole :many

-- "given a person, which portfolios are they part of?"
-- name: ListUserPortfolios :many
SELECT
    m.id                AS membership_id,
    m.created_at        AS joined_at,
    r.id                AS role_id,
    r.name              AS role_name,
    r.parent_role_id    AS role_parent_role_id,
    s.id                AS society_id,
    s.name              AS society_name,
    s.slug              AS society_slug
FROM memberships m
JOIN socroles r ON r.id = m.role_id
JOIN societies s ON s.id = r.soc_id
WHERE m.user_id = $1 AND m.ended_at IS NULL AND s.archived_at IS NULL
ORDER BY s.name, r.name;

-- TODO: GetUserMembershipInSociety :one

-- ends the current holders of a role. with user_id set, only that holder
-- name: EndRoleMemberships :many
UPDATE memberships
SET ended_at = now()
WHERE role_id = sqlc.arg('role_id')
  AND ended_at IS NULL
  AND (sqlc.narg('user_id')::uuid IS NULL OR user_id = sqlc.narg('user_id'))
RETURNING id;

-- TODO: DeleteMembership :exec