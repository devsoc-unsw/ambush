-- name: CreateAuditEntry :one

-- name: GetAuditEntry :one
SELECT * FROM audit_log
WHERE id = $1;

-- every row written by one user action, oldest first (replay order for undo)
-- name: ListChangeSet :many

-- name: ListAuditLog :many

-- name: ListAuditLogForEntity :many

-- name: GetUndoOfEntry :one