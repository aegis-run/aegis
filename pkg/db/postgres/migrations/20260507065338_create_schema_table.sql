-- +goose Up
CREATE DOMAIN "schema_hash" AS text
  CHECK (value ~ '^[0-9a-f]{64}$');

CREATE TABLE "schema" (
  pk         bigint        GENERATED ALWAYS AS IDENTITY,
  hash       "schema_hash" NOT NULL,
  data       bytea         NOT NULL,
  written_at xid8          NOT NULL DEFAULT pg_current_xact_id(),
  created_at timestamptz   NOT NULL DEFAULT now(),

  CONSTRAINT schema_pk          PRIMARY KEY (pk),
  CONSTRAINT schema_hash_unique UNIQUE (hash),
  CONSTRAINT schema_data_check  CHECK (octet_length(data) > 0)
);

CREATE INDEX schema_created_at_idx ON schema (created_at DESC);

-- +goose Down
DROP INDEX schema_created_at_idx;

DROP TABLE "schema";
DROP DOMAIN "schema_hash";
