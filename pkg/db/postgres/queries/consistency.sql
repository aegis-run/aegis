-- name: GetCurrentXactID :one
SELECT pg_current_xact_id()::xid8;

-- name: GetSnapshotXmax :one
SELECT pg_snapshot_xmax(pg_current_snapshot())::xid8;

-- name: IsXactVisible :one
SELECT pg_visible_in_snapshot(@xid::xid8, pg_current_snapshot())::boolean;
