# go-pg-router — Project Milestones

Each milestone answers one question: **"Does this core thing work?"**
Do not move to the next milestone until the current one's acceptance test passes.

---

## M1 — TCP Listener (Echo Server)

**Question answered:** Can we accept a TCP connection?

- Listen on `localhost:5432`
- Accept a connection, read bytes, echo them back, close
- **Acceptance test:** `psql` connects and you see the bytes echoed (it will fail at the protocol level, but data flows)
- **What's faked:** Everything except TCP itself

---

## M2 — Parse PostgreSQL Startup Message

**Question answered:** Can we speak the beginning of the PostgreSQL protocol?

- Read and decode the startup message (length + protocol version + key-value pairs)
- Send back a fake `AuthenticationOk` + `ReadyForQuery`
- **Acceptance test:** `psql` connects and reaches the prompt (even though no real Postgres is behind it)
- **What's faked:** The entire backend — we pretend to be Postgres ourselves

---

## M3 — Full Proxy to a Single Backend

**Question answered:** Can we forward a real connection to a real Postgres?

- Connect to a real Postgres on startup
- Pipe bytes: client → Postgres and Postgres → client
- **Acceptance test:** `psql` connects through the router, runs `SELECT 1`, gets a result
- **What's faked:** Nothing — first milestone with a real Postgres

---

## M4 — SQL Parsing & Read/Write Classification

**Question answered:** Can we tell a read from a write?

- Intercept the Query message (`'Q'` type byte)
- Parse the SQL with `pganalyze/pg_query_go`
- Return `"read"` or `"write"` classification
- **Acceptance test:** Unit tests covering `SELECT`, `INSERT`, `UPDATE`, `DELETE`, `CREATE TABLE`, transactions
- **What's faked:** Network — pure logic, tested without any connections

---

## M5 — Read/Write Routing to Two Backends

**Question answered:** Does routing actually work end to end?

- Config has two endpoints: `primary` and `replica`
- Reads go to replica connection, writes go to primary connection
- **Acceptance test:** Run two local Postgres instances (Docker Compose); verify writes land on primary, reads on replica
- **What's faked:** Nothing — real two-node setup locally

---

## M6 — Concurrent Connections

**Question answered:** Can multiple clients use the router simultaneously?

- Each accepted connection gets its own goroutine
- No shared mutable state between connections
- **Acceptance test:** 10 clients connect simultaneously, all get correct results with no data races (`go test -race`)
- **What's faked:** Load — simulated with concurrent test clients

---

## M7 — Configuration Loading

**Question answered:** Is it configurable without recompiling?

- Load primary/replica addresses from env vars or a config file (YAML/TOML)
- Validate config at startup, fail fast with a clear error message
- **Acceptance test:** Change env var, restart, behavior changes accordingly
- **What's faked:** Nothing new

---

## M8 — Docker Image

**Question answered:** Can we ship it?

- Write a `Dockerfile` (multi-stage build: compile Go binary → copy into minimal image)
- Build and push to a container registry (GitHub Container Registry is free)
- **Acceptance test:** `docker run` the image, it starts and accepts connections
- **Unlocks:** M9 (CI/CD pipeline can now publish real images)

---

## M9 — GitHub Actions CI/CD Pipeline

**Question answered:** Is every commit verified automatically?

> Set this up early — ideally right after M1 passes, so CI has something real to test from the start.

Pipeline stages:
1. `lint` — run `golangci-lint`
2. `test` — run `go test -race ./...`
3. `build` — compile the binary
4. `push` — build Docker image and push to registry (on `main` branch only)

- **Acceptance test:** A bad commit fails the pipeline; a good commit produces a published image automatically

---

## M10 — Kubernetes Integration Test

**Question answered:** Does it actually work as a sidecar in a real cluster?

- Write a k8s manifest: a Pod with the app container + go-pg-router sidecar
- Use a local cluster (`kind` or `minikube`) in CI
- Deploy Postgres primary + replica as Pods/Services
- Run a test Job that connects through the sidecar and verifies routing
- **Acceptance test:** The k8s test Job completes successfully in CI

---

## Suggested Build Order

```
M1 → M2 → M3 → M4 → M5 → M6 → M7 → M8 → M10
          ↑
     Add M9 (CI) here — after M1,
     so every subsequent milestone is automatically verified
```
