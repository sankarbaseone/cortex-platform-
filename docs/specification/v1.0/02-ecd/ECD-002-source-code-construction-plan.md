# ECD-002 — Source Code Construction Plan

**Level:** 2 (extends RFC-012 O.2, RFC-014 G.0/G.1, RFC-002, RFC-005, RFC-007). ECD-001 tree is the frame; this document specifies files WITHIN units.
**Method (deterministic, no "similar/etc."):** §1 defines the **Normative Service File Set (NSFS)** — the exact file list every Go service MUST contain, with per-file contracts. §2–§6 then specify only the files that are unit-specific (domain logic), by name, for every unit in repository.yaml. A unit's complete file list = NSFS ⊕ its unit-specific table. This composition is exhaustive: if a file is not in NSFS and not in the unit table, it does not exist in V1.0.

## 1. NSFS — every Go service (16 services)
Per file: purpose · public interface · LOC range · logging/metrics/tracing · errors · concurrency · tests required.
| File | Contract |
|---|---|
| `cmd/<name>/main.go` | wiring only: config load→validate→print(redacted), otel init, adapters construct, healthz/readyz mount, graceful shutdown (SIGTERM drain ≤25s, matches terminationGracePeriod 30s). 60–120 LOC. No business logic (lint: forbidden imports of internal/domain funcs beyond constructor). Test: none (covered by itest boot). |
| `internal/config/config.go` | typed struct, env tags (`envconfig`), `Validate() error` (fail-fast), `Redacted() string`. Every field documented with default+range. 80–200 LOC. Unit tests: table of invalid configs. |
| `internal/domain/*.go` | pure logic per unit tables below. No IO, injected `Clock`, errors as typed sentinels wrapped `%w`. Concurrency: pure funcs goroutine-safe; stateful aggregates documented per type (ECD-003). Unit tests mandatory, diff-cov ≥80%. |
| `internal/ports/ports.go` | ALL interfaces the domain needs (repo, bus, clock, flags). One file unless >300 LOC then split by consumer. Mocks generated. |
| `internal/adapters/pg/repo.go` (+`queries.sql` if sqlc) | implements ports; sqlc-generated queries (chosen over ORM: compile-checked SQL, zero reflection; rejected GORM). RLS GUC set per call via `nyxpg.WithTenant(ctx)`. Integration tests against testcontainer PG with RLS enabled (cross-tenant read must fail — required test per RFC-009 I.6). |
| `internal/adapters/kafka/{producer.go,consumer.go}` | via `libs/go/nyxbus`: envelope encode/decode, schema-registry check, retry/DLQ per RFC-005 B.4, idempotency (strategy declared in unit table: `upsert` or `dedup-table`). Consumer concurrency: one goroutine per partition, ordered keys honored. Required tests: duplicate-delivery, out-of-order, poison-pill→DLQ. |
| `internal/adapters/grpc/server.go` | generated svc impl → domain calls; interceptors from `nyxgrpc` (auth, tenant, deadline, otel, recovery, ratelimit — order fixed exactly that). |
| `internal/adapters/http/` | only in control-plane-api (grpc-gateway there); other services have no HTTP business endpoints — /metrics,/healthz served by `nyxhttp.Ops` listener :9090. |
| `internal/adapters/redis/cache.go` | only where unit table says; keys per RFC-007 C.5 keyspace. |
| `migrations/NNNN_*.up.sql` | per RFC-011 J.5 expand-contract; each PR’s migration reviewed with dry-run output. |
| `service.yaml` | machine manifest: SLOs (RFC-014 G.3), HPA signal, flags, dashboards, alerts — genmanifests consumes. |
| `Dockerfile` | distroless static; nonroot; readonly fs (RFC-014 G.0). 20 LOC fixed template. |
| `dashboards/service.json`, `alerts/rules.yaml`, `runbook.md` | required by RFC-011 Q observability gate; alert↔runbook link check. |
Cross-cutting numbers (apply to all): logs JSON schema `{ts,level,svc,tenant,trace_id,msg,...}` via `nyxotel.Logger`; RED metrics auto from interceptors, names `nydux_<svc>_<subsystem>_...`; every outbound call wrapped `nyxhttp/nyxgrpc.Breaker` (5-fail→open 30s); memory requests=limits per ECD-014 budgets (Run 3/4 — interim values in service.yaml now).

## 2. Unit-specific files — compiler pipeline (the wedge)
### analyzers/compiler-analyzer (Rust)
| File | Purpose · key items · LOC · perf · tests |
|---|---|
| `crates/canonicalizer/src/lib.rs` | RFC-002 §2.4 exact rules: strip debug/comments, SSA alpha-rename in def order, sort commutative pairs (op table `COMMUTATIVE_OPS` enumerated in ECD-003), constant normalization, arch/version embed. Pub: `fn canonicalize(ir: &Ir, meta: &Meta) -> CanonicalIr`. 600–900 LOC. p95 <20ms/kernel @10k-instr. Tests: golden fixtures per parser-matrix version; property: idempotence `canon(canon(x))==canon(x)`; rename-invariance. |
| `crates/canonicalizer/src/hash.rs` | `fn kernel_hash(&CanonicalIr)->[u8;32]` BLAKE3; `family_hash` masks tunable-param groups (mask table in ECD-003). Property: tile-size change ⇒ same family, diff kernel. |
| `crates/ptx-parser/src/{lexer.rs,ast.rs,parser.rs,features.rs}` | PTX ISA 7.x–8.x grammar; outputs `PtxFeatures{regs,instr_mix,mem_ops,barriers}`. 1500–2500 LOC. Fuzz target `fuzz/ptx.rs` (mandatory, RFC-002 2.9). Unknown ISA ⇒ `Err(UnsupportedVersion)` → MEASUREMENT_ONLY path (never panic; fuzz asserts). |
| `crates/sass-decoder/src/{opcodes_sm80.rs,opcodes_sm90.rs,decode.rs,confidence.rs}` | text-SASS decode; per-arch opcode tables as data (const arrays); emits `SassFeatures{spills,occ_inputs,sched_hints}` + `decode_confidence` (fraction of recognized ops). OQ-01 honored: confidence <0.6 ⇒ static KES capped per RFC-002 2.5. |
| `crates/kes/src/{components.rs,score.rs,profiles.rs,calibrate.rs}` | components exactly RFC-002 formulas; `profiles.rs` = weight tables {training,prefill,decode,hpc} (numeric values in ECD-003 §3.9 — fixed defaults, refit path via calibrate); `score.rs` geometric combine + confidence; `calibrate.rs` offline weight-fit CLI (regression vs measured throughput). Property tests: monotonicity (any component ↓ ⇒ KES ↓), bounds [0,100], reproducibility stddev<2 (RFC-002 2.9). |
| `crates/ncu-join/src/lib.rs` | Nsight-Compute CSV ingest → counter map → component measurement; per-driver normalization table (data file `norm/<driver>.toml`). |
| `crates/cache/src/lib.rs` | Redis hash-cache keyed `cache:kes:{hash}:{kes_model_version}`; invalidation via pubsub channel `kes.model.bump`. |
| `src/main.rs`, `src/pipeline.rs`, `src/grpc.rs`, `src/sandbox.rs` | job loop (NATS local queue), pipeline per RFC-002 §2.2 mermaid, gVisor/no-net enforcement check at boot (refuse start if sandbox absent in prod profile — RFC-009 I.7). Concurrency: worker pool = min(cores, queue-configured), each analysis single-threaded (parsers not Sync by design; simpler ownership). |

### services/kernel-registry (Go) — unit files
`internal/domain/{kernel.go,toolchain.go,register.go}`: upsert semantics (first_seen/last_seen), status transitions SCORED↔STATIC↔MEAS_ONLY (legal transition table in ECD-003), toolchain fingerprint = SHA-256 of sorted component versions. Outbox pattern (RFC-012 O.4): `internal/adapters/pg/outbox.go` + relay goroutine (poll 250ms, batch 500, exactly-once via row lock+offset column). Consumes: none. Produces: `compiler.kernel.*` per RFC-005. Idempotency: upsert-by-natural-key.

### services/regression-svc — unit files
`internal/domain/{cri.go,prescreen.go,gate.go,noise.go}`: CRI math verbatim RFC-002 §2.6 (σ_noise from historical run variance, Welford online); `prescreen.go` logistic model inference (weights loaded from model registry blob, version pinned in config); `gate.go` evaluates expressions `CRI>0.10` (grammar: `CRI(>|>=)float` only — no general DSL in V1; ASM-002-1). Produces `compiler.regression.detected`; serves `POST /regressions/checks` via api. Tests: golden regression corpus (30 labeled cases) FP<10% asserted in CI (RFC-010 L.2).

### services/bench-runner — unit files
`internal/domain/{plan.go,variance.go,fingerprint.go}`: run plan n≥5 median (RFC-002 2.6), variance guard refuses node if MAD/median>3% (RFC-014), env fingerprint = {driver,clockcaps,toolchain,gpu_uuid,thermal_state}. Executes via K8s Job in DP with pinned pod spec (charts/nydux-collector ships template). Verify mode: numerical-equivalence harness (abs/rel tol per dtype table ECD-003) then perf compare — implements RFC-008 correctness gate + RFC-013 #18.

## 3. Unit-specific files — graph/finance/twin/policy/audit (per service, domain files only)
- **graph-svc:** `domain/{ontology.go,writer.go,queries.go,privacy.go}` — writer is SOLE mutation path (Cypher templates as consts, one per edge type); `queries.go` named queries Q_REG_BLAST/Q_COST_TO_KERNEL/Q_REC_PRIOR with per-query timeout + materialized-view fallback (RFC-003 D.10); `privacy.go` global-graph allowlist re-check (defense-in-depth behind edge-gateway). Consumers (upsert idempotency): kernel.scored, rec.*, cost.calculated, regression.detected.
- **finance-svc:** `domain/{rates.go,slices.go,attribution.go,anomaly.go}` — time-weighted apportionment; committed-spend deconvolution (algorithm: proportional-fill of commit pool by on-demand-equivalent usage, remainder at commit rate — exact formula ECD-004 Run 2); parked-slice on missing rate (never guessed, RFC-014).
- **savings-svc:** `domain/{baseline.go,ledger.go,shapley.go,reanchor.go}` — frozen_state hash-anchored; Shapley exact ≤10 actions else MC m=200 (RFC-004 4.3); reanchor single-flight redlock; replay-bundle exporter (`export.go`) with model versions+seeds.
- **twin-svc:** `domain/{roofline.go,collective.go,pipeline.go,residual.go,scenario.go}` — analytical core formulas verbatim RFC-004 4.1; residual LightGBM inference via ONNX runtime (training offline in `tools/`); LOW_SUPPORT via kNN distance τ (τ value in model metadata).
- **policy-svc:** `domain/{eval.go,toolchainstate.go}` — OPA embedded (rego lib), decision cache 5s TTL, fail-closed for block / fail-open+audit for warn (RFC-014 table verbatim).
- **audit-svc:** `domain/{chain.go,verify.go,anchor.go}` — entry_hash=BLAKE3(prev||canonical(entry)) with canonical JSON (RFC-000 conventions: sorted keys, no floats in entries — amounts as integer micros; ASM-002-2), gap detector consumer, daily head anchor to object-lock bucket (RFC-009 I.8).

## 4. Unit-specific files — DP Rust (collector, edge-gateway, runtime-analyzer)
- **collector:** `src/{dcgm.rs,nvml.rs,k8swatch.rs,relabel.rs,sampler.rs,spool.rs,clock.rs}` — dcgm via dcgm-exporter scrape localhost (chosen over FFI: stability; rejected direct libdcgm FFI for V1), pod↔GPU relabel per RFC-001 A.3 failure-mode fix; spool = segmented ring (segments 64MiB, fsync per segment, 24h default); clock guard skew>2s refuse (RFC-005 B.7).
- **edge-gateway:** `src/{ingest.rs,privacy.rs,spool.rs,mtls.rs}` — privacy.rs implements RFC-003 D.9 allowlist as compiled schema (build fails if proto adds field not classified allow/deny — the CI privacy test); reject = hard error metric `egress_filter_reject_total` + alert (RFC-014).
- **runtime-analyzer:** `src/{vllm.rs,sglang.rs,trtllm.rs,nccl.rs,traces.rs}` — scrapers normalize to `runtime.serving.metrics`; nccl profiler-plugin reader.

## 5. Unit-specific — agents, web, cli, operator, sinks
- **agents:** files exactly per ECD-001 tree; `orchestrator/{lifecycle.py,budget.py,killswitch.py}`; each `domains/<agent>/agent.py` = declarative spec {tools_scope, triggers, judge_extras} consumed by generic loop (RFC-008 H.7 table transcribed to code constants in ECD-003).
- **web/app:** one feature dir per RFC-010 K.3 screen (13 dirs, names: dashboard,kernels,compilers,graph,cost,savings,simulate,recs,governance,agents,reports,admin,notifications); shared `components/` (KesRadar, RooflinePlot, CriHeatmap, EvidenceViewer, AuditChainBrowser); Playwright per primary flow (K.4).
- **cli/nydux:** `cmd/{login,scan,kernel,kernels,regressions,simulate,savings,policy}.go` + exit codes 0/1/2 (RFC-006 F.7).
- **operator:** `controllers/{cluster,collector,policy,baseline}_controller.go`, CRDs under `api/v1alpha1/`.
- **ch-sink/ts-sink:** consumer→batch insert (CH: async_insert + keeper offsets exactly-once; TS: COPY batches 5k rows/500ms whichever first); late-row % metric (RFC-005 B.7).

## 6. Config, logging, metrics naming (fixed vocabularies)
Env prefix `NYDUX_<SVC>_...`; durations Go syntax; every service exposes `NYDUX_<SVC>_LOG_LEVEL`, `_OTLP_ENDPOINT`, `_FLAGS_SOURCE`. Metric namespaces: `nydux_ingest_*, nydux_kes_*, nydux_cri_*, nydux_graph_*, nydux_fin_*, nydux_sav_*, nydux_twin_*, nydux_policy_*, nydux_audit_*, nydux_agent_*` — genmanifests builds metrics-registry.yaml (Run 4) from code annotations `// metric:`.

## Assumptions & conflicts raised
- ASM-002-1: gate expression grammar limited to CRI comparisons (RFC-006 shows only that form). ASM-002-2: audit canonical JSON forbids floats (money as integer micros) to make hashing deterministic — extends RFC-009 I.8. ASM-002-3: sqlc + golang-migrate chosen (RFC named migrate only); rejected: ent/GORM (reflection, RLS friction).
- **RFC_CONFLICT-003:** RFC-002 §2.3 says MLIR parsers "vendored per Triton tag" (Rust) while MLIR bindings are C++/Python. ECD resolves: Triton IR captured as TEXT via SDK hook; Rust parses the stable textual dialect subset needed for features; full-fidelity path uses PyO3 bridge module `crates/pybridge` invoking pinned triton package inside the sandbox. Both paths behind one trait `IrFrontend`. Confirm acceptability.
