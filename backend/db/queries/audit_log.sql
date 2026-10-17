-- name: CreateAuditEntry :one
INSERT INTO audit_log (
    change_set_id, org_id, entity_type, entity_id,
    action, actor_id, before, after, undoes_id
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetAuditEntry :one
SELECT * FROM audit_log
WHERE id = $1;

-- every row written by one user action, oldest first (replay order for undo)
-- TODO: ListChangeSet :many

-- TODO: ListAuditLog :many

-- TODO: ListAuditLogForEntity :many

-- TODO: GetUndoOfEntry :one
