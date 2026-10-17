-- dev dummy data, safe to re-run (ON CONFLICT skips existing rows)
--
-- DevSoc
-- └── ExecTeam (leader: Co-President)  Co-President x2, VP Internals,
--     │                                  VP External, VP Projects
--     └── Warchest (leader: Director)  Director, Subcom x3
-- CSESoc
-- └── Exec (leader: President)         President
--
-- users are fixed uuids: ...0001 alice, 0002 bob, 0003 carol, 0004 dan,
-- 0005 eve, 0006 frank, 0007 grace, 0008 heidi

INSERT INTO orgs (name, slug) VALUES
    ('DevSoc', 'devsoc'),
    ('CSESoc', 'csesoc')
ON CONFLICT (slug) DO NOTHING;

-- root teams
INSERT INTO teams (org_id, name)
SELECT o.id, t.name
FROM (VALUES ('devsoc', 'ExecTeam'), ('csesoc', 'Exec')) AS t(slug, name)
JOIN orgs o ON o.slug = t.slug
ON CONFLICT (org_id, name) DO NOTHING;

-- Warchest sits under ExecTeam
INSERT INTO teams (org_id, parent_id, name)
SELECT p.org_id, p.id, 'Warchest'
FROM teams p
JOIN orgs o ON o.id = p.org_id
WHERE o.slug = 'devsoc' AND p.name = 'ExecTeam'
ON CONFLICT (org_id, name) DO NOTHING;

INSERT INTO roles (org_id, team_id, name)
SELECT t.org_id, t.id, r.name
FROM (VALUES
    ('devsoc', 'ExecTeam', 'Co-President'),
    ('devsoc', 'ExecTeam', 'VP Internals'),
    ('devsoc', 'ExecTeam', 'VP External'),
    ('devsoc', 'ExecTeam', 'VP Projects'),
    ('devsoc', 'Warchest', 'Director'),
    ('devsoc', 'Warchest', 'Subcom'),
    ('csesoc', 'Exec', 'President')
) AS r(slug, team, name)
JOIN orgs o ON o.slug = r.slug
JOIN teams t ON t.org_id = o.id AND t.name = r.team
ON CONFLICT (team_id, name) DO NOTHING;

-- leader slots, set after the roles exist
UPDATE teams t
SET leader_role_id = r.id
FROM (VALUES
    ('devsoc', 'ExecTeam', 'Co-President'),
    ('devsoc', 'Warchest', 'Director'),
    ('csesoc', 'Exec', 'President')
) AS l(slug, team, role)
JOIN orgs o ON o.slug = l.slug
JOIN roles r ON r.org_id = o.id AND r.name = l.role
WHERE t.org_id = o.id AND t.name = l.team;

INSERT INTO memberships (role_id, user_id, term)
SELECT r.id, m.user_id::uuid, '2026'
FROM (VALUES
    ('devsoc', 'Co-President', '00000000-0000-0000-0000-000000000001'),
    ('devsoc', 'Co-President', '00000000-0000-0000-0000-000000000002'),
    ('devsoc', 'VP Projects',  '00000000-0000-0000-0000-000000000003'),
    ('devsoc', 'Director',     '00000000-0000-0000-0000-000000000004'),
    ('devsoc', 'Subcom',       '00000000-0000-0000-0000-000000000005'),
    ('devsoc', 'Subcom',       '00000000-0000-0000-0000-000000000006'),
    ('devsoc', 'Subcom',       '00000000-0000-0000-0000-000000000007'),
    ('csesoc', 'President',    '00000000-0000-0000-0000-000000000008'),
    -- alice is in both orgs, so portfolios show more than one org
    ('csesoc', 'President',    '00000000-0000-0000-0000-000000000001')
) AS m(slug, role, user_id)
JOIN orgs o ON o.slug = m.slug
JOIN roles r ON r.org_id = o.id AND r.name = m.role
ON CONFLICT (role_id, user_id) WHERE ended_at IS NULL DO NOTHING;
