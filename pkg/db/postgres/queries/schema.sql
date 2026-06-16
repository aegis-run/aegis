-- name: InsertSchemaVersion :one
INSERT INTO "schema" (
  hash, data
) VALUES (
  @hash, @data
)
ON CONFLICT (hash) DO NOTHING
RETURNING hash, written_at, created_at;

-- name: GetLatestSchemaVersion :one
SELECT
  s.hash,
  s.data,
  s.written_at,
  s.created_at
FROM "schema" s
ORDER BY s.pk DESC
LIMIT 1;

-- name: GetSchemaVersionByHash :one
SELECT
  s.hash,
  s.data,
  s.written_at,
  s.created_at
FROM "schema" s
WHERE s.hash = @hash;
