-- dev dummy data
-- safe to re-run
-- (ON CONFLICT skips existing rows)

INSERT INTO societies (name, slug) VALUES
    ('Seed Society', 'seed-society'),
    ('Other Society', 'other-society')
ON CONFLICT (slug) DO NOTHING;

INSERT INTO socroles (soc_id, name)
SELECT id, 'Exec' FROM societies WHERE slug = 'seed-society'
ON CONFLICT (soc_id, name) DO NOTHING;

INSERT INTO socroles (soc_id, parent_role_id, name)
SELECT s.id, r.id, 'Projects Team'
FROM societies s
JOIN socroles r ON r.soc_id = s.id AND r.name = 'Exec'
WHERE s.slug = 'seed-society'
ON CONFLICT (soc_id, name) DO NOTHING;

INSERT INTO socroles (soc_id, name)
SELECT id, 'Member' FROM societies WHERE slug = 'other-society'
ON CONFLICT (soc_id, name) DO NOTHING;

INSERT INTO memberships (role_id, user_id)
SELECT r.id, '00000000-0000-0000-0000-000000000001'
FROM socroles r
JOIN societies s ON s.id = r.soc_id
WHERE (s.slug, r.name) IN (('seed-society', 'Projects Team'),
                            ('other-society', 'Member'))
ON CONFLICT (role_id, user_id) DO NOTHING;
