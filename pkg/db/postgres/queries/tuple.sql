-- name: InsertTuple :one
INSERT INTO "tuple" (
  resource_type, resource_id, relation, subject_type, subject_id, subject_permission
) VALUES (
  @resource_type, @resource_id, @relation, @subject_type, @subject_id, @subject_permission
)
ON CONFLICT (
  resource_type, resource_id, relation,
  subject_type, subject_id, subject_permission
) DO NOTHING
RETURNING pk;

-- name: DeleteTuple :exec
DELETE FROM "tuple"
WHERE resource_type = @resource_type
  AND resource_id = @resource_id
  AND relation = @relation
  AND subject_type = @subject_type
  AND subject_id = @subject_id
  AND subject_permission = @subject_permission;

-- name: FindTuples :many
SELECT * FROM "tuple"
WHERE (resource_type = @resource_type OR @resource_type = '')
  AND (resource_id = @resource_id OR @resource_id = '')
  AND (relation = @relation OR @relation = '')
  AND (subject_type = @subject_type OR @subject_type = '')
  AND (subject_id = @subject_id OR @subject_id = '')
  AND (subject_permission = @subject_permission OR @subject_permission = '')
  AND pk > @last_pk
ORDER BY pk ASC
LIMIT @limit_val;

-- name: FindTuplesByResourceBatch :many
SELECT * FROM "tuple"
WHERE relation = @relation
  AND resource_type = @resource_type
  AND resource_id = ANY(@resource_ids::text[]);
