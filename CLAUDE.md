# NYDUX Platform — Agent Instructions
- Read RFC-000..014 in /docs/rfcs before architectural work. RFCs are binding.
- Languages: Go 1.23 (services), Rust 1.80 (analyzers/), Python 3.12 (agents/, sdk), TS 5 (web/).
- NEVER: add cross-layer imports (see depguard.yaml), log payload contents, egress raw IR,
  write to graph outside graph-svc, bypass RLS, add a Kafka field without registry update.
- ALWAYS: UUIDv7 ids, tenant_id on every row/message, table-driven tests, context.Context first arg,
  errors wrapped with %w, feature-flag new user-visible behavior, update OpenAPI/proto with handler.
- Build: `make build` · Test: `make test` (unit) / `make itest` (integration, needs docker) ·
  Lint: `make lint` · Local stack: `make dev-up` (kind + PG/CH/Kafka/Redis + seed fixtures).
- One PR = one story ID (M-###). Conventional Commits. Diff coverage ≥80%.
