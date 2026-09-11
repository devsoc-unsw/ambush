-- CreateSocRole
INSERT INTO socroles (soc_id, parent_role_id, name)
VALUES ($1, $2, $3)
RETURNING *;

-- GetSocRole
SELECT * FROM socroles
WHERE id = $1;

-- ListSocRoles
-- Every role for a given society
SELECT * FROM socroles
WHERE soc_id = $1
ORDER BY parent_role_id NULLS FIRST, name;

-- UpdateSocRole
UPDATE socroles
SET name = $2, parent_role_id = $3, updated_at = now()
WHERE id = $1
RETURNING *;

-- ListRoleDescendants
WITH RECURSIVE descendants AS (
    SELECT r.id FROM socroles r WHERE r.id = $1
  UNION
    SELECT c.id
    FROM socroles c
    JOIN descendants d ON c.parent_role_id = d.id
)
SELECT id FROM descendants;

-- ListRoleAncestors
WITH RECURSIVE ancestors AS (
    SELECT r.id, r.parent_role_id, r.name, 0 AS depth
    FROM socroles r WHERE r.id = $1
  UNION ALL
    SELECT p.id, p.parent_role_id, p.name, a.depth + 1
    FROM socroles p
    JOIN ancestors a ON p.id = a.parent_role_id
    WHERE a.depth < 64
)
SELECT id, parent_role_id, name, depth FROM ancestors
ORDER BY depth;
