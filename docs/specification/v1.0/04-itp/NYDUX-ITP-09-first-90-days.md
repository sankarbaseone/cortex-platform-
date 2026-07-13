# First 90-Day Execution Plan (13 weeks) — optimized for earliest vertical slice

**W1 (S0):** T-001 repo-init + T-002 protos + T-005 CI live. Validation: empty-green pipeline. Infra: none (kind+compose only).
**W2 (S0→S1):** T-003 libs-core + T-004 migrations. Validation: RLS cross-tenant itest red-green. **Checkpoint: M0.**
**W3 (S1):** T-101 kernel-registry (exemplar → real) + T-104 tenant/auth start. Tests: F-KR specs, token SoD.
**W4 (S1):** T-103 gateway+collector MVP + T-102 sinks. Infra: staging Terraform apply (eks, pg, kafka, kms, buckets — modules shipped). Validation: e2e fixture device→CH on staging. **Checkpoint: M1 ingest spine. Demo #1: live GPU telemetry from NYDUX's own cluster in the raw UI shell.**
**W5 (S2):** T-201 canonicalizer+hash (goldens, idempotence). Parallel: NCU corpus capture on XE7740/R670 (IR-04 burn-down).
**W6 (S2):** T-202 parsers (fuzz 1h no-panic) + T-203 KES with corpus goldens.
**W7 (S2):** T-204 compiler-analyzer + minimal kernels screen. Validation: first-score p95 <5min on corpus. **Checkpoint: M2 VERTICAL SLICE. Demo #2 (the fundable demo): submit real Triton kernel → KES score visible with component breakdown.** Debt checkpoint #1: burn quarantine list to zero before S3.
**W8 (S2):** T-205 regression-svc + T-206 bench-runner. Validation: CLI gate exit codes vs staging. **Checkpoint: M3. Demo #3: pre-upgrade CRI risk report for two CUDA versions — the ThinkVaultAI/design-partner pitch artifact.**
**W9 (S3):** T-301 recommender + T-302 policy-svc (fail-closed itest) + T-303 audit-svc start.
**W10 (S3):** T-303 finish (1M-verify bench) + T-304 rec-lifecycle API. Validation: approve→apply→verify→rollback e2e; audit verify green. **Checkpoint: M4. Demo #4: full governed optimization loop with hash-chained audit — the enterprise-sales demo.** Alpha gate review; internal tenant goes live on staging permanently.
**W11 (S4):** T-401 finance (sum-preservation property) + T-305 graph-svc.
**W12 (S4):** T-402 twin (MAPE gate) + T-403 savings (replay bundle golden). **Checkpoint: M5. Demo #5: dual-report savings statement for a simulated month.** Debt checkpoint #2 + velocity recalibration for S5/S6 plan.
**W13:** buffer/overflow (reality tax) + Beta-gate prep: partner #1 onboarding dry-run (TTFHW clock), S5 sprint planning.

Standing weekly rhythm: Mon plan (build-order next tasks) · daily `make verify` green before new work · Fri dashboard regen + GAP/ADR queue + spend check. Exit position at Day 90: M0–M5 done, Alpha achieved (~W10), Beta gates in progress, S5 surface work starting with recalibrated velocity.
