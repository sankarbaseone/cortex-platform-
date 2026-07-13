# RFC-000 — NYDUX Engineering Specification V2: Index, Conventions, Open Questions

**Status:** Approved · **Supersedes:** none · **Extends:** NYDUX Master Spec V1 (frozen)
**Audience:** All engineering, product, QA, SRE, security, technical writing, and Claude Code.

## 0.1 Purpose
V1 defined WHAT NYDUX is and WHY. V2 defines HOW, to implementation-grade precision. No engineer should need to make an architectural decision not covered or explicitly delegated here. Where V2 silently disagrees with V1, V2 wins only if the RFC states "V1 correction" with rationale; otherwise V1 stands.

## 0.2 RFC Set
| RFC | Title | Owns (V2 Sections) |
|---|---|---|
| 001 | Overall Architecture | A (enterprise/logical/physical/deployment/HA/DR/network/cloud/on-prem/hybrid/edge/air-gap) |
| 002 | Compiler Intelligence Engine | E (KES, CRI, recommendation, RCA), Phase 6 implementation |
| 003 | Knowledge Graph Design | D (ontology, embeddings, traversal, GC, privacy) |
| 004 | Digital Twin & Simulation Engine | E (twin, simulation, savings attribution, forecasting, capacity) |
| 005 | Event Bus & Telemetry | B (full event catalog, schemas, retries, DLQ, ordering) |
| 006 | API & SDK Specification | F (REST/gRPC/WS/GraphQL, OpenAPI, protobuf, SDK, CLI) |
| 007 | Database Schema | C (Postgres/Timescale/ClickHouse/Graph/Vector/Redis/Blob) |
| 008 | AI Agent Framework | H (agents, memory, RAG, guardrails, prompts, MCP) |
| 009 | Security Architecture | I (SOC2/ISO/NIST/FedRAMP, zero trust, threat model, SBOM) |
| 010 | Frontend & UX | K, L (screens, flows, wireframes, PRD-level requirements) |
| 011 | CI/CD & Release Engineering | J, Q (repo, git, testing pyramid, quality gates, canary) |
| 012 | Claude Code Implementation Guide | O, P, M, N (folder structure, prompts, backlog, sprints) |
| 013 | Patent Engineering | R (all 22 V1 patents expanded) |
| 014 | Microservice Catalog | G (per-service contracts, SLOs, circuit breakers, flags) |

## 0.3 Conventions (binding on all RFCs)
- **Languages:** Go 1.23+ for control-plane services; Rust 1.80+ for collectors and IR/SASS parsers (perf + memory safety at the edge); Python 3.12 for SDK, analyzers that wrap LLVM/MLIR Python bindings, and agents; TypeScript 5.x + React 18 for frontend.
- **IDs:** UUIDv7 everywhere (time-ordered). Kernel identity = `kernel_hash` = BLAKE3-256 of canonicalized IR (RFC-002 §2.4).
- **Time:** UTC, RFC3339 in APIs, epoch-nanos internally.
- **Serialization:** protobuf on the wire (gRPC + Kafka), JSON at REST edges. Schema registry mandatory (RFC-005).
- **Errors:** RFC 9457 problem+json for REST; gRPC status codes with `ErrorInfo` detail.
- **Naming:** services `nydux-<domain>` (e.g., `nydux-compiler-analyzer`); topics `nydux.<layer>.<entity>.<event>.v<N>`; DB tables snake_case singular.
- **SemVer** for every API, schema, chart, SDK. Breaking change = major bump + 12-month dual-support window (enterprise LTS = 36 months per V1 Phase 5).
- **Tenancy:** every row, message, and object key carries `tenant_id` (UUIDv7). No cross-tenant read path exists except hash-level graph aggregation (RFC-003 §D.9).

## 0.4 Definition of "Done" for any RFC-scoped feature
Code + unit/integration tests (RFC-011 pyramid) + metrics/logs/traces wired + runbook entry + API docs + feature flag + security review checklist + benchmark (if hot path) merged behind flag, canaried, then GA.

## 0.5 SECTION S — Open Questions Register (canonical)
Each item: ID · Question · Why it matters · Owner-by-default · Decision deadline · Interim assumption.

**OQ-01 · SASS decode fidelity.** How far can we statically infer Blackwell/Rubin SASS scheduling without NVIDIA docs? *Matters:* KES accuracy. *Owner:* Compiler team. *Deadline:* before KES v1 GA. *Interim:* probabilistic model calibrated to Nsight Compute measurements only; static-only KES flagged `confidence<0.6`.
**OQ-02 · Graph DB choice.** Neo4j (mature Cypher, licensing $) vs JanusGraph (OSS, ops burden) vs Postgres+AGE (one less system). *Interim:* Postgres+Apache AGE for MVP; revisit at >50M edges (RFC-003 §D.11).
**OQ-03 · Savings counterfactual acceptance.** Will customers accept twin-based baselines contractually? *Interim:* dual-report (twin + naive trailing-30-day baseline); contract uses the lower.
**OQ-04 · Kafka vs Redpanda.** Ops cost of Kafka for small on-prem installs. *Interim:* Kafka API as the contract; Redpanda permitted as drop-in for single-node/air-gap profiles.
**OQ-05 · Triton IR stability.** TTIR/TTGIR change across Triton releases. *Interim:* version-pinned parser matrix (RFC-002 §2.9); unknown version ⇒ measurement-only mode (V1 Phase 4.2 failure mode).
**OQ-06 · Per-request GPU attribution in shared vLLM.** Exact per-token device attribution is not directly observable. *Interim:* proportional model (tokens × phase weights) documented in RFC-004 §4.7; error bound published.
**OQ-07 · AMD parity timeline.** rocprof/ROCm-SMI field parity with DCGM. *Interim:* NVIDIA GA first; AMD collectors behind `backend.amd` flag; KES cross-vendor normalization (patent #22) deferred to Y3.
**OQ-08 · FedRAMP.** Pursue only on ≥2 US-gov-adjacent design partners. *Interim:* build to NIST 800-53 moderate control mapping now (RFC-009) so the delta is paperwork.
**OQ-09 · Agent write-actions liability.** Legal review of autonomous production mutations per jurisdiction. *Interim:* all write actions human-approved (V1 Phase 8 law); no exceptions in V1.0.
**OQ-10 · GraphQL surface.** Is GraphQL worth the resolver complexity vs REST+expand params? *Interim:* GraphQL read-only, frontend-BFF only (RFC-006 §F.5); not a public contract.
**OQ-11 · Embedding model for kernel similarity.** Off-the-shelf code embedder vs custom IR-token model. *Interim:* GraphCodeBERT-class embedder on canonical IR text; custom model when graph >1M kernels.
**OQ-12 · ClickHouse tenancy.** Shared cluster w/ row policies vs cluster-per-large-tenant. *Interim:* shared + row policies; dedicated for tenants >5TB/mo or contractual isolation.
**OQ-13 · Baseline drift.** How to handle model-architecture changes mid-savings-period? *Interim:* baseline re-anchoring event (RFC-004 §4.6.3) requiring customer co-sign.
**OQ-14 · License scanning of customer kernels.** Do we ever persist raw customer IR in SaaS control plane? *Decision (made):* Never. Hash + features only. Raw IR stays in-tenant (V1 Phase 5 deployment model).
**OQ-15 · Frontend real-time transport.** WS vs SSE. *Interim:* SSE for dashboards (simpler proxies), WS only for Graph Explorer interactive traversal.

## 0.6 Assumption Ledger (V2 additions to V1 Phase 2)
- A9: DCGM 10s cadence is sufficient ground truth for KES calibration windows ≥5 min. Confidence: High.
- A10: 95% of analyzable kernels in target customers originate from Triton/Inductor/XLA/cuBLAS-cuDNN; hand-written CUDA is long tail handled by SASS-only scoring. Confidence: Medium-High.
- A11: A single in-tenant analyzer pod (8 vCPU/32GB) sustains ≥50 kernel analyses/min at p95 <2s each (hash-cache hit rate ≥70% after week 1). Must be validated in Sprint 3 benchmark.
