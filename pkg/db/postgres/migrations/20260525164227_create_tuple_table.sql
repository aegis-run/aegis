-- +goose Up
CREATE DOMAIN type_name AS text
  CHECK (value ~ '^[a-z]([a-z0-9_]{0,62}[a-z0-9])?$');

CREATE DOMAIN relation_name AS text
  CHECK (value ~ '^[a-z]([a-z0-9_]{0,62}[a-z0-9])?$');

CREATE DOMAIN instance_id AS text
  CHECK (char_length(value) >= 1 AND char_length(value) <= 128);

CREATE DOMAIN subject_permission AS text
  CHECK (value = '' OR value ~ '^[a-z]([a-z0-9_]{0,62}[a-z0-9])?$');

CREATE TABLE "tuple" (
  pk                 bigint             GENERATED ALWAYS AS IDENTITY,
  resource_type      type_name          NOT NULL,
  resource_id        instance_id        NOT NULL,
  relation           relation_name      NOT NULL,
  subject_type       type_name          NOT NULL,
  subject_id         instance_id        NOT NULL,
  subject_permission subject_permission NOT NULL DEFAULT '',

  CONSTRAINT tuple_pk PRIMARY KEY (pk),
  CONSTRAINT tuple_unique UNIQUE (
    resource_type, resource_id, relation,
    subject_type, subject_id, subject_permission
  )
);

CREATE INDEX tuple_resource_idx ON tuple (resource_type, resource_id, relation);
CREATE INDEX tuple_subject_idx ON tuple (subject_type, subject_id);

-- +goose Down
DROP INDEX tuple_subject_idx;
DROP INDEX tuple_resource_idx;
DROP TABLE "tuple";
DROP DOMAIN subject_permission;
DROP DOMAIN instance_id;
DROP DOMAIN relation_name;
DROP DOMAIN type_name;
