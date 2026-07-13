# RFC-012 — Claude Code Implementation Guide

**Status:** Approved · **Owns Sections O, P, M, N.** This RFC is written so Claude Code (or any engineer) can begin implementation immediately.

## O.1 Canonical CLAUDE.md (place at repo root; excerpted normative content)
```md
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
```

## O.2 Folder structure (matches RFC-011 J.1; service-internal layout)
```
services/<name>/
  cmd/<name>/main.go        # wiring only
  internal/domain/          # pure logic, no IO — unit-test heavy
  internal/ports/           # interfaces (repo, bus, clock)
  internal/adapters/{pg,ch,kafka,redis,http,grpc}/
  internal/config/          # env-parsed struct, validated at boot
  api/                      # generated stubs (do not edit)
  migrations/
  dashboards/ alerts/ runbook.md
```
Pattern: hexagonal (ports/adapters). Domain never imports adapters. Generated code never edited.

## O.3 Naming & style
Go: golangci defaults + depguard; packages singular nouns; interfaces `-er`. Rust: clippy pedantic on parsers. Python: ruff strict, mypy strict, pydantic models for all IO. TS: eslint strict, no `any`. Proto: buf style. SQL: snake_case, no `select *` in app code.

## O.4 Design patterns (use) / anti-patterns (reject)
Use: hexagonal services; outbox pattern for DB-write+event atomicity (pg outbox table → Kafka relay); upsert-by-natural-key consumers; repository interfaces per store; typed feature flags; table-driven tests; golden-file tests for parsers.
Reject: shared DB between services (each service owns its schema; cross-service reads via API/events only — ClickHouse/Timescale are designated shared read stores with view contracts); distributed transactions; sagas where an upsert suffices; god-clients; mocks of types you own instead of fakes; time.Now() in domain (inject clock).

## O.5 Testing strategy for Claude Code
Every task ends with: unit tests written FIRST for domain logic; integration test if adapter touched; idempotency test if consumer touched; fixture added to golden corpus if parser touched; `make test itest lint` green before PR.

## O.6 Architecture rules (machine-enforced)
`tools/depguard.yaml` encodes: web→bff→api only; api→services via grpc clients only; services/*→libs allowed; agents→MCP tools only (no direct DB); analyzers→edge-gateway only for egress. CI fails on violation (RFC-011 Q).

## P — Developer Prompt Library (reusable Claude Code prompts; parameters in ⟨⟩)
- **P1 New endpoint:** "Implement `⟨METHOD PATH⟩` per RFC-006 F.4 in service ⟨name⟩. Add proto + grpc-gateway annotation, handler in adapters/http, domain use-case in internal/domain with interface in ports, PG repo adapter, migration if schema change (expand-contract per RFC-011 J.5), RSQL filters ⟨list⟩, cursor pagination, problem+json errors, unit + integration tests, OpenAPI regenerated, dashboard panel + alert if new failure mode. Story M-⟨id⟩."
- **P2 New event consumer:** "Add consumer for `nydux.⟨topic⟩` in ⟨service⟩: envelope decode, schema-registry check, idempotent upsert by ⟨natural key⟩, retry/DLQ wiring per RFC-005 B.4, idempotency integration test (duplicate + out-of-order delivery), lag metric + alert."
- **P3 Migration:** "Create expand-contract migration for ⟨change⟩: expand migration, backfill batched 10k rows with progress log, contract migration gated by flag ⟨flag⟩, rollback notes, dry-run output in PR description."
- **P4 New UI screen:** "Build ⟨screen⟩ per RFC-010 K.3 item ⟨n⟩: route, Query hooks against BFF, loading/empty/error states, AA accessibility (axe test), SSE subscription ⟨channel⟩ if live, Playwright e2e for primary flow."
- **P5 Unit tests:** "Write table-driven tests for ⟨pkg⟩ covering: happy path, each documented edge case in RFC-⟨n⟩ §⟨s⟩, error wrapping, property tests where math (use rapid/hypothesis)."
- **P6 Optimize query:** "Profile ⟨query⟩ with EXPLAIN ANALYZE at 10× current rows (seed via tools/seed), propose index or CH ORDER BY change, add regression benchmark, document tradeoff in service README."
- **P7 New agent tool:** "Expose ⟨capability⟩ as MCP tool per RFC-008 H.5: JSON schema, server-side validation, scope ⟨read/write⟩, audit args-hash, golden-task eval additions, deny-test proving scope enforcement."
- **P8 Runbook:** "Write RB-⟨AREA⟩-⟨NNN⟩ per RFC-011 J.9 for alert ⟨name⟩ with diagnosis tree from failure modes in RFC-⟨n⟩."
- **P9 Parser support:** "Add ⟨artifact⟩ version ⟨v⟩ to parser matrix per RFC-002 2.3: grammar update, golden fixtures, fuzz corpus seeds, MEASUREMENT_ONLY fallback test, matrix doc row."
- **P10 Refactor:** "Refactor ⟨module⟩ preserving public behavior: characterization tests first, then change, benchmarks ±5% proof, no API change without RFC note."

## M — Engineering Backlog (Epic→Story with estimates; P0 = MVP)
**E1 Ingest & Infra layer (P0)** — M-101 collector DaemonSet Rust w/ DCGM (5d) · M-102 spool+backfill (3d) · M-103 edge-gateway mTLS gRPC (5d) · M-104 CH sink + rollups (4d) · M-105 k8s pod↔GPU relabel (3d).
**E2 Compiler engine (P0)** — M-201 canonicalizer+hash (5d) · M-202 Triton/FX capture SDK hook (5d) · M-203 PTX parser (8d) · M-204 SASS feature decoder sm80/90 (10d, risk-high OQ-01) · M-205 static KES (5d) · M-206 NCU join + dynamic KES (5d) · M-207 kernel registry svc+API (4d) · M-208 hash cache (2d).
**E3 Regression (P0)** — M-301 bench-runner (5d) · M-302 CRI math + storage (4d) · M-303 static pre-screen classifier (5d) · M-304 CLI CI gate + JUnit (3d).
**E4 Recommendations (P1)** — M-401 pattern lib 1–5 (8d) · M-402 rec svc + inbox API (4d) · M-403 approval/audit flow (4d) · M-404 verify-runner (4d).
**E5 Graph (P1)** — M-501 PG+AGE setup + graph-svc (5d) · M-502 event consumers→graph (4d) · M-503 named queries Q_* (4d) · M-504 embeddings+similar (5d) · M-505 egress privacy filter + CI test (3d).
**E6 Finance (P1)** — M-601 rates+cost slices (5d) · M-602 attribution incl. kernel-$ (6d) · M-603 anomaly (4d).
**E7 Twin/Savings (P2)** — M-701 analytical core (8d) · M-702 residual model (6d) · M-703 scenarios API+UI (6d) · M-704 baseline+savings ledger (6d) · M-705 Shapley attribution (4d).
**E8 Platform (P0)** — M-801 auth/tenancy/RLS (6d) · M-802 policy-svc OPA (4d) · M-803 audit chain (4d) · M-804 control-plane-api skeleton (4d) · M-805 helm+operator (6d) · M-806 CI/CD per RFC-011 (6d).
**E9 Frontend (P1)** — M-901 shell+auth (4d) · M-902 Dashboard (5d) · M-903 Kernel Explorer (6d) · M-904 Compiler Explorer (5d) · M-905 Rec Inbox (5d) · M-906 Cost Explorer (5d).
**E10 Agents (P2)** — M-1001 orchestrator+MCP (6d) · M-1002 Judge+guardrails (5d) · M-1003 Advisor (4d) · M-1004 evals harness (4d).
Dependencies: E2←E1(gateway), E3←E2, E4←E2/E5, E6←E1, E7←E5/E6, E9←respective APIs, E10←E5. Risk-high: M-204, M-702, M-705.

## N — Sprint plan to V1.0 (2-week sprints, small team + Claude Code; MVP=GA of L.1/L.2 + platform)
- **S1:** M-801, M-804, M-806, repo scaffold, dev-up stack. Exit: authed hello-API deployed to staging via full pipeline.
- **S2:** M-101, M-103, M-104. Exit: real DCGM samples visible in CH from a test cluster.
- **S3:** M-201, M-202, M-207, benchmark A11 validation. Exit: kernels registered with hashes end-to-end.
- **S4:** M-203, M-205, M-208. Exit: static KES on golden corpus; scores in API.
- **S5:** M-204 (part), M-206, M-105. Exit: dynamic KES with NCU on lab cluster; per-pod attribution.
- **S6:** M-301, M-302, M-304. Exit: CI gate demo — CUDA bump blocked on golden regression.
- **S7:** M-901–M-903, M-102. Exit: Kernel Explorer usable by design partner.
- **S8:** M-501–M-503, M-802, M-803. Exit: graph queries power "kernel-$" view; audit verify passes.
- **S9:** M-601, M-602, M-904, M-505. Exit: Cost Explorer + pre-upgrade risk report; privacy CI test green.
- **S10:** M-401(patterns1–3), M-402, M-905. Exit: first real recommendation approved by design partner.
- **S11:** M-403, M-404, M-303. Exit: applied→verified loop closes; RESULTED_IN edges appear.
- **S12:** hardening, SOC2 evidence automation, docs, airgap-sim rehearsal, load test 10k-GPU synthetic. Exit: **V1.0 release candidate** → canary per RFC-011 J.6.
Each sprint's DoD = RFC-000 §0.4. Files/modules per story follow O.2 layout; every story maps to a prompt in P.
