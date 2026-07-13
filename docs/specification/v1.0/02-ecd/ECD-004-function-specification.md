# ECD-004 — Function Specification

**Level:** 2 · Extends RFC-002/004/014, ECD-002/003 (frozen). **Artifact-first:** exemplar code in `artifacts/go-examples/` is normative and compilable; this document is the index + the specs Claude Code implements with prompt P1–P10.

## 4.1 Dependency Injection — PLATFORM STANDARD (mandated decision)
**Chosen:** manual constructor injection, wired exclusively in `cmd/<name>/main.go` via `nyxrun.Group`.
**Why:** the dependency graph must stay visible and depguard-enforceable; 16 services × ~6 deps each is trivially hand-wireable; failures are compile errors, not runtime container errors.
**Rejected:** google/wire (codegen step, zero benefit at this scale), uber/fx (runtime graph, lifecycle magic conflicts with crash-only design), service locators (hides deps).
**Wiring contract:** order = config → otel → stores → outbox/bus → repos → domain use-cases → transports → `nyxrun.Group(...).RunUntilSignal(25s)`. The exemplar `main.go` (artifacts) is the template every service copies.

## 4.2 Exemplar package (normative, compilable)
`artifacts/go-examples/kernelregistry_exemplar.go` + `_test.go` — covers Repository, Service (use-case), Consumer, Producer(outbox), DI main, and the full test-first set (unit table, failure, property, benchmark, security). Every other service replicates these shapes; deviations are ECD violations.

## 4.3 Function specification format (binding)
Every public function gets a spec ID `F-<SVC>-NN` in its package doc, containing: signature · inputs/outputs · error conditions (typed sentinels) · business rules · validation · concurrency · perf target · security validation · metrics/logs/traces emitted · tests required. The exemplar demonstrates F-KR-01 fully. Specs for V1.0 public functions follow (condensed rows carry every required field; pseudo-code only where the artifact isn't already real code).

### Compiler layer
| ID | Function | Contract highlights |
|---|---|---|
| F-KR-01 | `Registrar.RecordScored(ctx,Kernel) error` | fully specified in exemplar (rules: hash regex, transition table, first_seen preservation; p95<5ms; errors ErrInvalidHash/ErrIllegalTransition; metric `nydux_kes_records_total{status}`; finding F-KR-01a: add empty-tenant guard — approved addition, not architecture change) |
| F-KR-02 | `Registrar.RecordCompiled` | same shape; toolchain fp validated 64-hex; upsert toolchain row if new (approval=unreviewed) |
| F-RG-01 | `CRI.Compare(family,from,to,runs) (Comparison,error)` | Δperf=(t2−t1)/t1 medians n≥5; regressed iff Δ>max(0.03,2σ_noise); σ via Welford from history; errors: ErrInsufficientRuns(n<5), ErrShapeMismatch; pure func, goroutine-safe; p95<1ms; property: antisymmetry Compare(a,b).Δ≈−Compare(b,a).Δ/(1+Δ) |
| F-RG-02 | `Fleet.CRI(window) (FleetCRI,error)` | Σ contrib over time-weighted top-95% families; excludes unbaselined (listed separately); metric `nydux_cri_value` gauge |
| F-RG-03 | `Gate.Eval(expr,FleetCRI) (bool,error)` | grammar `^CRI>=?float$` only (ASM-002-1); exit-code mapping done in CLI |
| F-BR-01 | `Runner.Bench(plan) (RunStats,error)` | n≥5, median+MAD; refuses node if MAD/median>3% → ErrNoisyNode; env fingerprint recorded; verify mode adds numerical-equivalence per ECD-003 §3.6 tolerance table before perf compare (RFC_CONFLICT-001 resolution: one service, job kinds bench|verify) |
| F-RC-01 | `Detector.Scan(kernelCtx) ([]RecDraft,error)` | pattern lib 1–12 (closed set); every draft carries ≥1 EvidenceRef (proto-validated); gain priors from Q_REC_PRIOR; confidence from prior support; risk per RFC-004 4.10 formula |

### Graph / finance / savings / twin
| ID | Function | Contract highlights |
|---|---|---|
| F-GS-01 | `Writer.Apply(event) error` | SOLE mutation path; Cypher template per edge type; idempotent MERGE by natural key; consumer-lag metric; failure: parse→DLQ |
| F-GS-02 | `Queries.RegBlast(from,to)` | 2-hop; timeout 500ms; fallback materialized view `mv_reg_blast` (refreshed hourly); returns freshness watermark |
| F-GS-03 | `Queries.CostToKernel(team,window)` | join time_share×CostSlice; returns completeness ratio when slices parked |
| F-FN-01 | `Attr.Apportion(slices,usage) ([]AttributionRow,error)` | time-weighted; commit deconvolution: fill commit pool by on-demand-equivalent usage proportion `share_i = odeq_i/Σodeq`, pool remainder billed at commit rate to unallocated bucket (formula final; extends RFC-007 note); property: Σ rows = Σ input usd exactly (integer micros) |
| F-SV-01 | `Savings.Compute(period,baseline) (SavingsPeriod,error)` | S=Σ demand·(unitB−unitA) floored per line; dual-report; contractual=min; negative shown; determinism: same inputs+model versions ⇒ byte-identical replay bundle |
| F-SV-02 | `Shapley.Attribute(actions,twin)` | exact ≤10 else MC m=200 with CI; property: efficiency Σφ=total ± integer-rounding ≤ |A| micros |
| F-TW-01 | `Twin.Predict(scenario) (Prediction,error)` | analytical core (roofline+αβ+1F1B bubbles) then residual multiply; LOW_SUPPORT when kNN dist>τ; p95<100ms/scenario; golden-scenario MAPE≤15% asserted in CI |

### Policy / audit / auth (security-critical)
| ID | Function | Contract highlights |
|---|---|---|
| F-PL-01 | `Engine.Decide(policy,input) (Decision,error)` | OPA eval; cache 5s; **fail-closed** for block, fail-open+audit for warn (exact per RFC-014); p95<50ms; every decision → Kafka + ledger ref |
| F-AU-01 | `Chain.Append(entry) (seq,hash,error)` | hash=BLAKE3(prev‖canonical); canonical JSON: sorted keys, NFC, no floats; single-writer per tenant partition; gap ⇒ Sev-1 |
| F-AU-02 | `Chain.Verify(tenant,fromSeq)` | streaming recompute; 1M entries ≤10s (RFC-010 L.5) — benchmark required |
| F-AT-01 | `Token.Mint/Verify` (approval tokens) | Ed25519, claims{user,rec_id,exp≤24h,scope}; verify checks not-author (separation of duties) |

### Analyzer (Rust; signatures normative)
`canonicalize(&Ir,&Meta)->Result<CanonicalIr,CanonError>` · `kernel_hash(&CanonicalIr)->[u8;32]` · `parse_ptx(&str)->Result<PtxFeatures,PtxError>` (never panics; fuzz-proven) · `decode_sass(&str,Arch)->SassFeatures` (total function; confidence encodes unknowns) · `score(KesInput)->Score` (monotonic, bounded, reproducible stddev<2 — property tests in crate).

## 4.4 Test-first & DoD (applies to every F-spec)
Unit table + failure + property (where math) + benchmark (hot path) + integration (adapter touched) + security test (isolation/authz where relevant); diff-cov ≥80%; spec ID referenced in test names; DoD = RFC-000 §0.4 + spec row complete.

## 4.5 Claude Code Task Graph (execution DAG; feeds ECD-015)
```
EPIC E2 Compiler engine
 └─ CAP kernel scoring
    ├─ S M-201 canonicalizer     files: crates/canonicalizer/*  funcs: canonicalize,kernel_hash
    │    tests: golden+property  build: make test-rs  verify: idempotence prop green
    │    commit: "feat(canon): canonicalizer per RFC-002 2.4 [M-201]"  merge: gates Q  release: train
    ├─ S M-205 static KES        depends: M-201  files: crates/kes/*  funcs: score,profiles
    │    verify: monotonicity+bounds props; weights == ECD-003 §3.5 table
    ├─ S M-207 registry svc      depends: M-205  files: services/kernel-registry/** (exemplar template)
    │    verify: itest RLS-cross-tenant-fails; consumer duplicate/out-of-order tests
    └─ S M-30x regression chain  depends: M-207 ... (per RFC-012 N sprint order, unchanged)
```
Same expansion pattern applies to every epic E1–E10; the machine-readable full DAG ships as `build-order.yaml` in Run 4 (per the ECD brief, AI-native manifests are Run-4 scope; task-graph *format* is fixed here so Run 4 is mechanical).
