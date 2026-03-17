# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**go-pg-router** is a lightweight Kubernetes sidecar that automatically routes PostgreSQL queries based on operation type — reads go to replicas, writes go to the primary — with no application code changes required.

## Commands

Since the project has no `go.mod` or source files yet, initialize Go module setup with:

```bash
go mod init github.com/tinymurky/go-pg-router
```

Once the project is set up, standard Go commands apply:

```bash
go build ./...          # Build all packages
go test ./...           # Run all tests
go test ./... -run TestName   # Run a single test
go vet ./...            # Static analysis
```

## Coding style

This project follow belowing coding guide line.

1. revive rule that store in ./revive.toml
2. [uber-go/guide](https://github.com/uber-go/guide)

## Architecture Intent

The router acts as a PostgreSQL protocol proxy (sidecar) that:
1. Intercepts incoming PostgreSQL wire protocol connections from the application
2. Parses SQL queries to determine operation type (read vs. write)
3. Routes read queries (`SELECT`) to replica connections and write queries (`INSERT`, `UPDATE`, `DELETE`, DDL, etc.) to the primary connection
4. Returns results transparently to the client

### Reference Projects

- [jhunt/pgrouter](https://github.com/jhunt/pgrouter) — similar PostgreSQL router implementation
- [pg-sharding/spqr](https://github.com/pg-sharding/spqr) — PostgreSQL shard/query router
- [pganalyze/pg_query_go](https://github.com/pganalyze/pg_query_go) — Go bindings for the PostgreSQL query parser (likely dependency for SQL parsing)
- [cockroachDB sql parser](https://github.com/cockroachdb/cockroach/tree/master/pkg/sql/parser) — alternative SQL parser reference
