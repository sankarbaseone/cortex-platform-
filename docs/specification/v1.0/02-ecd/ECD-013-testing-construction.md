# ECD-013 — Testing Construction
**Level:** 2 · Extends RFC-011 J.4 (frozen). **Artifacts are the spec.**
- `artifacts/testing/test-matrix.yaml` — canonical tier matrix: gates, targets, harnesses, and REQUIRED named tests per tier (RLS cross-tenant-fails, outbox exactly-once-effect, duplicate/out-of-order consumers, fail-closed policy, golden sets incl. byte-identical savings replay and twin MAPE gate); mutation testing scope + thresholds; flake policy with release-blocking rate.
- `artifacts/testing/k6-api-load.js` — executable load test asserting ECD-014 budgets as k6 thresholds (not observed — asserted).
- `artifacts/testing/chaos-plan.yaml` — 8 chaos experiments with explicit expectations (fail-closed approve during policy partition, strict-order survival on broker kill, spool/replay on link blackhole, audit-seq continuity on PG failover); any miss = Sev-2 GA blocker.
- Test data: `tools/seed` fixtures are the single corpus for dev/e2e/staging (NCU fixture kernels, seeded regressions, e2e SoD rec) — already normative via Playwright suite and golden sets.
- Coverage: diff ≥80% (CI), mutation ≥70% killed on the four math-critical packages at GA, e2e = shipped Playwright suite + smoke, a11y = axe gate.
