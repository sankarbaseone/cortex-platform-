# ECD-008 — API Construction Documents
**Level:** 2 · Extends RFC-006, ECD-003 §3.7/§3.8 (frozen). **Artifacts are the spec.**

## Shipped artifacts (canonical)
- `artifacts/api/openapi.yaml` — **complete OpenAPI 3.1**: 37 paths, 50 schemas, validated. Covers the entire RFC-006 F.4 endpoint map: kernels (incl. IR endpoint with SaaS-404-by-design), regressions + CI-gate check, toolchains + approval governance, full recommendation lifecycle (approve/reject/apply with idempotency keys, SoD, approval/rollback tokens), simulations (202/job), capacity plans, finance (attribution/savings/baselines/re-anchor with 423-locked), clusters/devices/rates, policies + decisions, audit entries + chain verify, agent tasks, tenants/users/roles/apikeys (secret-once), jobs.
- `artifacts/registries/api-registry.yaml` — machine contract: per-operation owner service, minimum role, audit flag, SoD flag, async flag, API-key eligibility, and the closed RSQL whitelists (eliminates the ECD-003 §3.8 interim). CI drift check: openapi.yaml ↔ registry mismatch fails.

## Construction rules already fixed elsewhere and inherited unchanged
Auth (RFC-006 F.2), pagination/filtering/errors/idempotency/rate limits/timeouts (F.3), gRPC style + interceptor order (F.5, ECD-002), protobuf sources (ECD-007 artifacts + transcription-eliminated packages below).

## Transcription elimination (per Run-3 directive)
The remaining proto packages (`infra/v1`, `runtime/v1`, `graph/v1`, `policy/v1`, `audit/v1`, `twin/v1`, `agent/v1`, `gateway/v1`) are now FULLY DERIVABLE with zero decisions: field lists = ECD-003 §3.2/§3.3 verbatim; message style = the four shipped .proto files; service methods = the api-registry owner mapping (each REST operation's gRPC twin via grpc-gateway annotation). Claude Code generates them mechanically in Sprint 1 (task in build-order, Run 4); any field not present in ECD-003 or the OpenAPI schemas DOES NOT EXIST.

## GraphQL & WebSocket
GraphQL: internal BFF read-only (OQ-10) — schema is generated from the OpenAPI read models 1:1 (no bespoke types); resolvers proxy REST list/get with DataLoader batching. WS: single endpoint `/v1/graph/ws`; messages `{op: expand|collapse|path, node_id}` → server pushes `{nodes[], edges[]}`; auth via first-message bearer; idle timeout 5 min.

## SDK & CLI examples (normative usage, matching shipped spec)
```python
cli.recommendations.approve(rec_id, rationale="verified on staging")   # → DecideResult.approval_token
cli.recommendations.apply(rec_id, approval_token=tok)                  # → rollback_token
cli.finance.attribution(dim="kernel", from_=t0, to=t1)                 # completeness surfaced
```
CLI `nydux regressions --from-fp F --to-fp T --fail-on 'CRI>0.10' --junit out.xml` maps to `checkRegressions`; exit 2 on gate_passed=false.

## Caching
GET responses: `Cache-Control: private, max-age=15` for list endpoints, `no-store` for audit/apikeys; ETag on kernel score (hash+kes_model_version) enabling conditional GET; server-side Redis rollup cache per RFC-007 C.5 only.
