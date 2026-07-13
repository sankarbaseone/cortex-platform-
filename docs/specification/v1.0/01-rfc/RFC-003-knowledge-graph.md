# RFC-003 — Knowledge Graph Design

**Status:** Approved · **Extends:** V1 Phase 4.7, Phase 12 (data moat) · **Owns Section D** entirely.

## D.1 Purpose
The graph joins cost↔kernel↔compiler↔hardware causality (the moat). Two graphs exist: **Tenant Graph** (full fidelity, in tenant scope) and **Global Graph** (hash-level, privacy-preserving aggregation across tenants). Same ontology; different node property whitelists (D.9).

## D.2 Storage (OQ-02 decision)
MVP: **Postgres + Apache AGE** (Cypher over PG) — one fewer stateful system, PITR for free, row-level security reuse. Migration trigger: >50M edges OR p95 3-hop traversal >500ms ⇒ move to JanusGraph/Neo4j behind the same `graph-svc` gRPC interface (repository pattern isolates store). Vector index: **pgvector** alongside (same DB) for kernel embeddings; migrate to Qdrant at >5M vectors.

## D.3 Ontology — Node types
| Label | Key | Core properties |
|---|---|---|
| KernelFamily | family_hash | op_class, pattern_tags[] |
| KernelVariant | kernel_hash | arch, ir_features(jsonb), kes, kes_components, kes_model_version, status |
| Toolchain | (name, version) | cuda_ver, ptxas_ver, triton_ver, torch_ver, host_cc |
| GpuSku | sku | peak_flops{fp32,bf16,fp8}, peak_bw, smem, regs, l2, ridge_points |
| Model | model_id | family, params, arch_summary |
| Workload | workload_id | phase(train/prefill/decode), parallelism{tp,pp,dp,ep} |
| Cluster | cluster_id | topology_ref, provider |
| Team | team_id | tenant-scoped |
| CostSlice | slice_id | window, usd, rate_basis |
| Recommendation | rec_id | pattern_id, gain_est{p50,p90}, confidence, state |
| RegressionEvent | reg_id | Δperf, mechanism_tags[], cri_contrib |
| MeasurementRun | run_id | source(ncu/dcgm/bench), stats |
| PolicyDecision | dec_id | policy_id, verdict, ledger_ref |

## D.4 Edge types
`(KernelVariant)-[:MEMBER_OF]->(KernelFamily)`
`(KernelVariant)-[:COMPILED_BY {flags}]->(Toolchain)`
`(KernelVariant)-[:MEASURED_ON {runs}]->(GpuSku)` via `(:MeasurementRun)`
`(Workload)-[:EXECUTES {time_share}]->(KernelVariant)`
`(Model)-[:HAS_WORKLOAD]->(Workload)` · `(Workload)-[:RUNS_ON]->(Cluster)`
`(Team)-[:OWNS]->(Workload)` · `(CostSlice)-[:ATTRIBUTED_TO {frac}]->(KernelVariant|Workload|Team)`
`(RegressionEvent)-[:FROM]->(Toolchain)` `-[:TO]->(Toolchain)` `-[:AFFECTS]->(KernelFamily)`
`(Recommendation)-[:TARGETS]->(KernelVariant)` `-[:EVIDENCED_BY]->(MeasurementRun)` `-[:RESULTED_IN {gain_measured}]->(MeasurementRun)`
`(KernelVariant)-[:SIMILAR_TO {cos}]->(KernelVariant)` (materialized top-k=20 from embeddings, refreshed nightly)

## D.5 Properties, Indexes
Btree on every key; GIN on jsonb ir_features and pattern_tags; composite (tenant_id,label,key) — tenant_id on every node/edge row (RFC-000 conventions). AGE graphs partitioned per tenant (`graph_name = g_<tenant>`) + one `g_global`.

## D.6 Embeddings & Similarity (OQ-11)
Embed canonical IR text (post-canonicalization RFC-002 §2.4) with GraphCodeBERT-class model → 768-d, L2-normalized, stored pgvector `ivfflat (lists=1024)`. Similarity = cosine. Use: rec retrieval ("what fixed kernels like this"), dedup near-misses, cold-start gain priors. Re-embed on embedder version bump (`emb_ver` property); dual-write during migration.

## D.7 Traversal, Ranking, Reasoning
**Core queries (named, versioned, in `graph-svc`):**
- `Q_REG_BLAST(toolchain_from, toolchain_to)`: families with RegressionEvents between versions, weighted by tenant's time_share ⇒ pre-upgrade risk report. 2-hop; p95 target <200ms.
- `Q_COST_TO_KERNEL(team, window)`: Team→Workload→Kernel time_share × CostSlice ⇒ per-kernel dollars.
- `Q_REC_PRIOR(kernel_hash)`: SIMILAR_TO top-k → their Recommendations with RESULTED_IN gains ⇒ gain prior distribution (feeds RFC-002 §2.7 expected_gain).
**Recommendation ranking:** `score = E[gain_$] · confidence − effort_cost − risk_penalty`, where E[gain_$] = gain_pct_P50 × kernel's attributed $/period (Q_COST_TO_KERNEL); risk_penalty from policy tags (e.g., quantization on regulated model). Top-N per tenant surfaced; full math in `recommender` (RFC-014).
**Reasoning:** deterministic queries first; agent narration reads query results only (RFC-008 guardrail: agents cannot write graph facts, only Recommendation/annotation nodes via approval flow).

## D.8 Updates, Learning, GC
- Writes only via `graph-svc` consuming Kafka events (kernel.scored, rec.created, rec.verified, cost.attributed…) — single-writer per entity key ⇒ no lock contention; idempotent upserts by natural key.
- **Learning loops:** (1) verified rec outcomes update gain priors; (2) regression classifier retrained monthly from RegressionEvents; (3) twin residual model (RFC-004) reads MeasurementRun corpus.
- **GC:** MeasurementRuns >18mo compacted to stats-only nodes; SIMILAR_TO edges pruned to top-20; orphan KernelVariants (no execution 12mo, no rec) archived to blob (reloadable). Tenant offboarding: `g_<tenant>` dropped within 30d contractual; global-graph hash rows unlinked (no reverse mapping exists by construction).

## D.9 Privacy (binding, implements OQ-14 / patent #3)
Global graph ingests ONLY: kernel_hash, family_hash, arch, ir_features limited to a reviewed numeric whitelist (op counts, tile params, occupancy inputs), KES components, toolchain versions, gain outcomes. FORBIDDEN in global graph: tenant ids (replaced by salted HMAC per-tenant→random contributor_id), model names, shapes beyond log-bucket class, any string from customer source. Enforcement: schema-level allowlist in `edge-gateway` (egress filter) + CI test that fails on new fields without privacy review. Contributor k-anonymity: a (family,toolchain,arch) aggregate is queryable only when ≥3 distinct contributors.

## D.10 Failure Modes
Graph lag (Kafka consumer behind) ⇒ UI shows freshness watermark; queries never block ingest. AGE query regression ⇒ named queries have per-query timeouts + fallback to precomputed materialized views (Postgres tables refreshed hourly) for the 5 core dashboards.

## D.11 Migration Plan (when OQ-02 trigger fires)
Dual-write via graph-svc adapter → backfill export (graphml) → shadow-read compare 2 weeks → cutover flag `graph.backend=janus`. No API change (gRPC stable).
