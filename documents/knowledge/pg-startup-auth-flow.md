# PostgreSQL Startup & Authentication Flow

Reference: https://www.postgresql.org/docs/current/protocol-flow.html

---

## Overview

Before any SQL query (Simple Query or Extended Query) can be sent, every PostgreSQL
connection must complete a **startup phase**. This phase establishes the session and
authenticates the client.

As a proxy/sidecar, go-pg-router sits between the app and the real Postgres backend.
This means there are **two separate authentication conversations** happening — one on
each side of the router.

---

## Two-Sided Proxy Architecture

```
[App / psql]  ←── frontend conn ──→  [go-pg-router]  ←── backend conn ──→  [Postgres]
```

| Side | Direction | Who authenticates? |
|---|---|---|
| Frontend | App → go-pg-router | go-pg-router fakes the server |
| Backend | go-pg-router → Postgres | go-pg-router acts as the client |

---

## Full Startup Flow (Both Sides)

```
App                    go-pg-router              Postgres
 │                          │                       │
 │── StartupMessage ────────▶│                       │
 │◀─ AuthenticationOk ───────│                       │   ← go-pg-router fakes this
 │◀─ ReadyForQuery ──────────│                       │   ← go-pg-router fakes this
 │                          │                       │
 │                          │── StartupMessage ──────▶│
 │                          │◀─ AuthMD5 / SCRAM ──────│
 │                          │── Password ─────────────▶│
 │                          │◀─ AuthenticationOk ──────│
 │                          │◀─ ReadyForQuery ──────────│
 │                          │                       │
 │── Query ─────────────────▶│── Query ───────────────▶│
 │◀─ Result ─────────────────│◀─ Result ───────────────│
```

---

## Frontend Side: App → go-pg-router

### Strategy: always trust the client (skip authentication)

go-pg-router runs as a Kubernetes sidecar on `localhost` inside the same Pod as the
app. The app can only reach the router because they share a network namespace — there
is no external network exposure. Asking for a password adds zero security value.

**Flow:**
1. App sends `StartupMessage`
2. go-pg-router replies `AuthenticationOk` immediately (no password challenge)
3. go-pg-router replies `ReadyForQuery`
4. App is now ready to send queries

This is the same strategy pgbouncer uses in `trust` auth mode.

---

## Backend Side: go-pg-router → Postgres

### Strategy: authenticate using configured credentials

The real Postgres will challenge go-pg-router with an auth request. go-pg-router
must respond with the correct password from its configuration (env vars or config
file — see M7).

**Common auth methods Postgres may send:**
- `AuthenticationCleartextPassword` — simple, rarely used in production
- `AuthenticationMD5Password` — MD5 hash of password + salt
- `AuthenticationSASL` (SCRAM-SHA-256) — modern default in Postgres 10+

---

## Startup Message Format

The startup message is the only PostgreSQL message with **no type byte prefix**.
All other messages start with a 1-byte type identifier.

```
Bytes  Content
──────────────────────────────────────────────────
4      Total message length (including these 4 bytes), big-endian int32
4      Protocol version: 0x00030000 = version 3.0
N      Key-value pairs, each key and value null-terminated (\0)
1      Final null byte \0 to signal end of parameters
```

Example parameters sent by `psql`:
```
user\0tinymurky\0database\0mydb\0application_name\0psql\0\0
```

---

## Server Response Messages

### AuthenticationOk
Tells the client: "you are authenticated, no further auth needed."

```
Byte  Content
─────────────────────────────────────────────────
'R'   Message type
4     Length = 8 (big-endian int32)
4     Auth type = 0 (int32, 0 means OK)
```

### ReadyForQuery
Tells the client: "the server is ready to accept a query."

```
Byte  Content
─────────────────────────────────────────────────
'Z'   Message type
4     Length = 5 (big-endian int32)
1     Transaction status: 'I' = idle, 'T' = in transaction, 'E' = error
```

---

## Milestone Mapping

| Milestone | What gets implemented |
|---|---|
| M2 | Frontend side only: read StartupMessage, reply AuthOK + ReadyForQuery |
| M3 | Backend side: connect to real Postgres, handle backend auth |
| M7 | Backend credentials loaded from env vars / config file |
