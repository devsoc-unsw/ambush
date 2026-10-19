-- the inserts select the role, so nothing is inserted unless the role is in
-- that org and the role, its team and the org are all active. with
-- non_root_only, roles in the root (exec) team are skipped too, which is how
-- grpc clients are kept out of exec
-- name: CreateMembership :one
INSERT INTO memberships (role_id, user_id, term)
SELECT r.id, sqlc.arg('user_id'), sqlc.narg('term')::text
FROM roles r
JOIN teams t ON t.id = r.team_id AND t.archived_at IS NULL
JOIN orgs o ON o.id = r.org_id AND o.archived_at IS NULL
WHERE r.id = sqlc.arg('role_id')
  AND r.org_id = sqlc.arg('org_id')
  AND r.archived_at IS NULL
  AND (NOT sqlc.arg('non_root_only')::boolean OR t.parent_id IS NOT NULL)
RETURNING *;

-- when the insert above misses, says whether the role exists in that org but
-- sits in the root team, so callers get a clearer error than not found
-- name: IsRootTeamRole :one
SELECT (t.parent_id IS NULL)::boolean AS is_root
FROM roles r
JOIN teams t ON t.id = r.team_id
WHERE r.id = sqlc.arg('role_id') AND r.org_id = sqlc.arg('org_id');

-- TODO: GetMembership :one

-- name: CreateMemberships :many
INSERT INTO memberships (role_id, user_id)
SELECT r.id, u.user_id
FROM (
    SELECT unnest(sqlc.arg('role_ids')::uuid[]) AS role_id,
           unnest(sqlc.arg('user_ids')::uuid[]) AS user_id
) AS u
JOIN roles r ON r.id = u.role_id
    AND r.org_id = sqlc.arg('org_id')
    AND r.archived_at IS NULL
JOIN teams t ON t.id = r.team_id AND t.archived_at IS NULL
JOIN orgs o ON o.id = r.org_id AND o.archived_at IS NULL
-- the WHERE has to match the partial unique index for postgres to find it
ON CONFLICT (role_id, user_id) WHERE ended_at IS NULL DO NOTHING
RETURNING *;

-- TODO: ListMembershipsByOrg :many

-- TODO: ListMembershipsByRole :many

-- "given a person, which portfolios are they part of?"
-- name: ListUserPortfolios :many
SELECT
    m.id                AS membership_id,
    m.created_at        AS joined_at,
    r.id                AS role_id,
    r.name              AS role_name,
    t.id                AS team_id,
    t.name              AS team_name,
    o.id                AS org_id,
    o.name              AS org_name,
    o.slug              AS org_slug
FROM memberships m
JOIN roles r ON r.id = m.role_id AND r.archived_at IS NULL
JOIN teams t ON t.id = r.team_id AND t.archived_at IS NULL
JOIN orgs o ON o.id = r.org_id AND o.archived_at IS NULL
WHERE m.user_id = $1 AND m.ended_at IS NULL
ORDER BY o.name, t.name, r.name;

-- TODO: GetUserMembershipInOrg :one

-- ends the current holders of a role. with user_id set, only that holder
-- name: EndRoleMemberships :many
UPDATE memberships
SET ended_at = now()
WHERE role_id = sqlc.arg('role_id')
  AND ended_at IS NULL
  AND (sqlc.narg('user_id')::uuid IS NULL OR user_id = sqlc.narg('user_id'))
RETURNING id;

-- ends a membership. COALESCE keeps the original time, so archiving twice is
-- a no-op. joining through roles scopes it to the org in the path, and
-- archived orgs are read only. archived roles and teams are allowed, so their
-- leftover memberships can still be ended
-- with non_root_only, memberships in the root (exec) team are skipped, same as
-- CreateMembership
-- name: ArchiveMembership :one
UPDATE memberships m
SET ended_at = COALESCE(m.ended_at, now())
FROM roles r
JOIN teams t ON t.id = r.team_id
JOIN orgs o ON o.id = r.org_id AND o.archived_at IS NULL
WHERE m.id = sqlc.arg('membership_id')
  AND r.id = m.role_id
  AND r.org_id = sqlc.arg('org_id')
  AND (NOT sqlc.arg('non_root_only')::boolean OR t.parent_id IS NOT NULL)
RETURNING m.*;

-- loads a membership in an org (not archived) with whether it sits in the root
-- team, and locks it so a role change can't race another change to the same
-- membership. outside a transaction the lock is released straight away
-- name: LockMembershipInOrg :one
SELECT m.id, m.role_id, m.user_id, m.term, m.ended_at,
       (t.parent_id IS NULL)::boolean AS in_root_team
FROM memberships m
JOIN roles r ON r.id = m.role_id
JOIN teams t ON t.id = r.team_id
JOIN orgs o ON o.id = r.org_id AND o.archived_at IS NULL
WHERE m.id = sqlc.arg('membership_id') AND r.org_id = sqlc.arg('org_id')
FOR UPDATE OF m;
