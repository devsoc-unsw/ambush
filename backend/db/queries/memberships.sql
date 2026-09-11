-- CreateMembership
INSERT INTO memberships (role_id, user_id)
VALUES ($1, $2)
RETURNING *;

-- GetMembership
SELECT * FROM memberships
WHERE id = $1;

-- DeleteMembership
DELETE FROM memberships
WHERE id = $1
RETURNING *;

-- ListUserMemberships
SELECT
    m.id         AS membership_id,
    m.created_at AS joined_at,
    r.id         AS role_id,
    r.name       AS role_name,
    r.parent_role_id,
    s.id         AS society_id,
    s.name       AS society_name,
    s.slug       AS society_slug
FROM memberships m
JOIN socroles r ON r.id = m.role_id
JOIN societies s ON s.id = r.soc_id
WHERE m.user_id = $1
ORDER BY s.name, r.name;
