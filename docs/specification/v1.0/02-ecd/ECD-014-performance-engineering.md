# ECD-014 — Performance Engineering
**Level:** 2 · Extends RFC-014 G.3 (frozen). **Artifacts are the spec.**
- `artifacts/perf/performance-budgets.yaml` — measurable budgets for every subsystem (API, ingest, compiler, regression, graph, finance, twin, policy, audit, frontend, agents), fixed benchmark methodology (n≥5 median, MAD<3%, pinned pool), the regression rule (>10% = blocker), required scaling curves, and the 100k-GPU capacity model.
- Enforcement mapping: k6 thresholds (weekly loadtest job) assert API/ingest budgets; criterion benches assert Rust-crate budgets per PR; go-bench guards F-spec hot paths (F-KR-01, F-AU-02 1M-verify, F-PL-01); Web Vitals asserted in Playwright (LCP/INP via CDP); twin MAPE golden is a CI gate.
- GPU-adjacent budgets: collector <1% node CPU / <0.5% GPU perf impact (DCGM-only sampling, no CUPTI in steady state); bench-runner isolates on `nydux.ai/bench=true` pool — noisy-neighbor exclusion is structural, not statistical.
- Memory discipline: requests==limits for memory platform-wide (chart), RSS slope <1%/day asserted in 72h soak.
