# Definition of Done — module level (paste as PR checklist)

A module (service/crate/feature) is DONE only when ALL hold:
- [ ] **Code complete** — every F-spec for the module implemented; exemplar shape; no TODO/FIXME; depguard clean.
- [ ] **Tests complete** — all required tiers per test-matrix.yaml; diff-cov ≥80%; required named tests present; failure tests per error path.
- [ ] **Benchmarks complete** — every ECD-014 budget in module scope has an asserting bench (criterion/go-bench/k6 threshold), green.
- [ ] **Security review complete** — threat-model rows touching the module re-checked; authz on every endpoint per api-registry; no secret/PII in logs (lint green); OPA tests green if policies touched.
- [ ] **Documentation complete** — godoc on exported items; runbook updated if new alert; registries updated same-PR.
- [ ] **Performance validated** — budgets asserted (above) AND soak-relevant metrics (RSS slope, goroutines) flat in a 1h local soak for stateful workers.
- [ ] **Observability complete** — uniform interceptor metrics live; module-specific metrics from metrics-registry emitting; dashboard panel renders with real data; alert fires in a forced test.
- [ ] **Deployment validated** — helm release installs on kind via values overlay; probes green; NetworkPolicy verified (deny test); rollback exercised once.
