-- +goose Up
CREATE TYPE audit_entity AS ENUM ('org', 'team', 'role', 'membership');
CREATE TYPE audit_action AS ENUM ('create', 'update', 'delete');

CREATE TABLE audit_log (
    id             BIGSERIAL PRIMARY KEY,
    change_set_id  UUID NOT NULL,          -- groups one user action into one row-set
    org_id         UUID NOT NULL,
    entity_type    audit_entity NOT NULL,
    entity_id      UUID NOT NULL,
    action         audit_action NOT NULL,
    actor_id       UUID NOT NULL,
    before         JSONB,
    after          JSONB,
    undoes_id      BIGINT REFERENCES audit_log (id),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),

    CHECK (
        (action = 'create' AND before IS NULL     AND after IS NOT NULL) OR
        (action = 'update' AND before IS NOT NULL AND after IS NOT NULL) OR
        (action = 'delete' AND before IS NOT NULL AND after IS NULL)
    )
);

CREATE INDEX idx_audit_log_org       ON audit_log (org_id, id DESC);
CREATE INDEX idx_audit_log_entity    ON audit_log (entity_type, entity_id, id DESC);
CREATE INDEX idx_audit_log_changeset ON audit_log (change_set_id);
CREATE INDEX idx_audit_log_actor     ON audit_log (actor_id, id DESC);

-- an entry may be undone at most once. chains (undo-of-undo = redo) still allowed
CREATE UNIQUE INDEX idx_audit_log_undoes
    ON audit_log (undoes_id) WHERE undoes_id IS NOT NULL;

-- +goose StatementBegin
CREATE FUNCTION audit_log_is_append_only() RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'audit_log is append-only (attempted %)', TG_OP
        USING ERRCODE = 'restrict_violation';
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER trg_audit_log_append_only
    BEFORE UPDATE OR DELETE ON audit_log
    FOR EACH ROW EXECUTE FUNCTION audit_log_is_append_only();

-- +goose Down
DROP TRIGGER trg_audit_log_append_only ON audit_log;
DROP FUNCTION audit_log_is_append_only;
DROP TABLE audit_log;
DROP TYPE audit_action;
DROP TYPE audit_entity;