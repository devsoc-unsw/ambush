-- CreateSociety
INSERT INTO societies (name, slug)
VALUES ($1, $2)
RETURNING *;

-- GetSociety
SELECT * FROM societies
WHERE id = $1;

-- GetSocietyBySlug
SELECT * FROM societies
WHERE slug = $1;

-- UpdateSociety
UPDATE societies
SET name = $2, slug = $3, updated_at = now()
WHERE id = $1
RETURNING *;
