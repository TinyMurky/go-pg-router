# go-pg-router
go-pg-router is a lightweight Kubernetes sidecar that automatically routes Postgres queries to the correct connection — reads go to replicas, writes go to the primary — no code changes required.

# Reference

- [jhunt/pgrouter](https://github.com/jhunt/pgrouter)
- [pg-sharding/spqr](https://github.com/pg-sharding/spqr)
- [cockroachDB sql parser](https://github.com/cockroachdb/cockroach/tree/master/pkg/sql/parser)
- [pganalyze/pg_query_go](https://github.com/pganalyze/pg_query_go)
