# Aegis: Distributed Authorization Engine

_A high-performance, content-addressed implementation of the Zanzibar design._

---

Archival record for the Aegis core engine. It implements the runtime environment for
the Aegis Entity Schema (AES), providing recursive graph evaluation, content-addressed
policy integrity, and distributed request dispatching.

## Repository Structure

- `bench/` Performance evaluation and synthetic workload generation.
- `cmd/aegis/` Service entry points and datastore management.
- `internal/engine/` Recursive graph traversal and permission resolution.
- `internal/dispatch/` Request distribution, coalescing, and caching.
- `internal/datalayer/` Abstracted persistence interface.
- `pkg/db/` PostgreSQL and in-memory storage drivers.
- `pkg/schema/` Codecs and SHA-256 hashing for AES artifacts.

---

## Reproducibility and Procedures

Deterministic environment managed via Nix. Operations orchestrated through `just`.

### Environmental Integrity

```bash
nix develop
```

### Toolchain Operations

_Operation A: Infrastructure Bootstrap_

```bash
just up
```

_Operation B: Server Execution_

```bash
just run
```

_Operation C: Persistence Reset_

```bash
just nuke
```

### Integrity Verification

_Verification A: Functional Testing_

```bash
just test
```

_Verification B: Static Analysis_

```bash
just lint
```

_Verification C: Security Auditing_

```bash
just security
```

---

## Technical Reference

For the full implementation and empirical analysis, refer to the sibling repository
and the archival thesis:  
[`Aegis: A Centralized Authorization System Based on Google Zanzibar.`](https://github.com/aegis-run/thesis)

_Faculty of Mathematics and Computer Science, University of Bucharest._
