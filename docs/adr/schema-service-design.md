# ADR-001: Schema Service Design

## Context

The Schema Service is the boundary between the AES compiler (Control Plane) and the
evaluation engine (Data Plane). It must store compiled schema versions, serve them to the
engine during permission evaluation, and provide sufficient revision semantics to support
future consistency token integration. Several non-obvious design decisions were made
during specification that require explicit documentation of their rationale.

## Decision

### 1. Server-side hash computation

The SHA-256 hash of each schema version is computed server-side from the bytes actually
persisted, not supplied by the compiler.

#### Consequences

- The hash is guaranteed to be consistent with what is stored; a mismatch between the
  revision token and the stored content is structurally impossible
- The compiler does not know the revision token until it receives the write response,
  which is acceptable since the token is returned immediately
- Protobuf serialization is not guaranteed to be deterministic across implementations; a
  compiler-supplied hash computed over in-memory bytes could diverge from the hash of the
  bytes received and stored by the Go server

---

### 2. Content-addressed versioning via SHA-256

Schema versions are identified by the SHA-256 hash of their serialized bytes, making the
storage model content-addressed.

#### Consequences

- Identical schemas submitted multiple times produce the same hash, making writes
  naturally idempotent without requiring an explicit deduplication query
- The hash is a stable, meaningful cache key for the engine's in-memory schema cache;
  cache invalidation reduces to hash comparison
- The hash is the natural revision token for future consistency binding between schema
  versions and permission check evaluation
- Schema versions are immutable by definition; the same hash always refers to the same content

---

### 3. Idempotent writes

If `Write` is called with a schema whose hash already exists in the database, the service
returns the existing record without performing a write.

#### Consequences

- The compiler can safely retry failed or uncertain write operations without producing
  duplicate records or inconsistent state
- No additional deduplication logic is required at the compiler level
- The response is identical whether the schema was just written or already existed, which
  keeps the compiler's state machine simple

---

### 4. Indefinite version retention

All schema versions are retained indefinitely. No garbage collection or expiry policy is
applied.

#### Consequences

- Pinned reads by hash are always satisfiable; a `NOT_FOUND` response unambiguously
  indicates the hash was never written, not that it was collected
- Future consistency token integration can safely reference any historical schema version
  without risk of the token becoming invalid
- Storage growth is proportional to the number of distinct schemas written, which is
  expected to be low in practice: schemas change infrequently relative to tuple writes
- No GC window configuration is required, reducing operational complexity for the MVP

#### Options considered

- **Keep only the latest version:** simpler storage model but breaks pinned reads and
  makes future consistency binding impossible without additional infrastructure
- **Keep N versions:** requires a configurable GC window and introduces the possibility of
  `NOT_FOUND` on valid historical tokens, mirroring the `Snapshot Expired` failure mode
  SpiceDB exposes on `at_exact_snapshot`

---

### 5. Schema versioned as an atomic unit

The entire schema (all type definitions, relations, and permission expressions) is
versioned as a single artifact. There is no per-type or per-namespace granularity in
versioning.

#### Consequences

- The hash captures a complete, consistent snapshot of the authorization policy at a point
  in time
- Cross-type references within the schema are always internally consistent within a given
  version: there is no possibility of evaluating a permission using a relation defined in
  one schema version and a type defined in another
- Schema migrations that affect multiple types are atomic from the engine's perspective
- Partial schema updates are not supported - the compiler must always emit a complete schema

---

### 6. `fully_consistent` as the default read consistency

When no `Consistency` requirement is specified on a `Read` request, the service defaults
to `fully_consistent` rather than `minimize_latency`.

#### Consequences

- Schema reads are infrequent relative to tuple reads and permission checks - the latency
  cost of bypassing caching is negligible in practice
- A stale schema read during permission evaluation is categorically more dangerous than a
  stale tuple read: it can cause the engine to evaluate a permission expression that no
  longer reflects the current authorization policy
- Callers that explicitly require lower latency can opt into `minimize_latency` or
  `at_least_as_fresh`

#### Options considered

- **`minimize_latency` as default:** consistent with SpiceDB's default, appropriate when
  the cached schema is known to be fresh; rejected here because the engine has no
  independent mechanism to verify schema freshness without a consistency token, which is
  deferred

---

### 7. Dual hash representation (Hex vs Base64)

The SHA-256 hash is represented as a Hex string in the persistence layer (PostgreSQL) and
as a Base64 string in the API layer.

#### Rationale

- **Hex for DB:** Standardizes storage, allows for simple regex-based constraints in Postgres,
  and ensures compatibility with standard database tooling and logging.
- **Base64 for API:** Provides a more compact representation for the wire format, and is
  the conventional format for revision tokens in Zanzibar-inspired systems.

#### Consequences

- The `pkg/schema` package must provide bidirectional conversion between both formats.
- Developers must be aware of the context when interpreting a hash: strings of length 64 are
  Hex (internal), while strings of length 44 are Base64 (external).

---

## Advices

- When consistency token integration is implemented, the `ConsistencyToken` returned from
  `Write` should encode the PostgreSQL `xid8` transaction ID captured via
  `pg_current_xact_id()` immediately after commit, not before. The token is only
  meaningful if it reflects the transaction that actually persisted the schema.
- The engine's in-memory schema cache should be keyed by `SchemaHash`, not by `created_at`
  or a sequence number. Hash-keyed caching is safe across restarts and requires no cache
  warming logic - a cache miss on a known hash is always resolvable by a pinned `Read`.
- Do not expose schema blob bytes directly through the API. The `Schema` message is always
  returned as a structured protobuf, with serialization and deserialization handled
  internally. Leaking raw bytes would couple callers to the serialization format and
  undermine the IR contract.
- If `pg_dump`/`pg_restore` is ever used to migrate the database, schema hashes will
  remain valid since they are computed from content, not from PostgreSQL-internal
  identifiers like `xid8`. This is an advantage over SpiceDB's revision model, which
  requires a repair step after restore.
