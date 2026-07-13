# Implementation Risk Register (implementation-only; architecture risks closed by spec)

| ID | Risk | P | I | Mitigation | Detection | Owner |
|---|---|---|---|---|---|---|
| IR-01 | Solo-founder bandwidth: review+build+ops concurrently | H | H | strict build-order focus; parallel lanes only when stable; hire trigger at Beta | sprint velocity <60% plan 2 sprints running | founder |
| IR-02 | Claude Code drift from spec accumulating quietly | M | H | CLAUDE.md constitution; registry drift CI; contract paths self-review cooling period; weekly gap/adr queue review | drift-check failures; PR diffs touching frozen paths | founder |
| IR-03 | Rust↔PyO3 bridge (RFC_CONFLICT-003) integration friction in T-204 | M | M | bridge behind IrFrontend trait with contract tests both paths; fuzz corpus early (T-202) | T-204 verify failures; MEASUREMENT_ONLY spike on dev corpus | founder |
| IR-04 | NCU fixture corpus insufficient → KES goldens weak | M | H | build corpus during S1 on NYDUX's own XE7740/R670 cluster before T-203 needs it | golden coverage report < profile matrix | founder |
| IR-05 | Budget misses discovered late (e.g. verify-1M ≤10s, first-score ≤5min) | M | M | budgets asserted per-task, not at GA; two-failure stop rule escalates early | red benches in task CI | founder |
| IR-06 | Dev/staging cloud spend creep during S2/S6 (GPU bench, CH) | M | M | scale-to-zero bench pool (TF shipped); staging schedules; weekly cost review vs ₹ budget | cloud billing alerts | founder |
| IR-07 | Design-partner unavailability delays Beta proofs (TTFHW, replay bundle) | M | M | line up 2 partners during Alpha (existing pipeline); internal tenant as fallback proof | Beta gate slip 2 weeks | founder |
| IR-08 | LLM provider instability breaks agent tasks in S5 | L | M | provider failover config; agents flag default-off; advisor is non-critical-path | RB-AG dashboards | founder |
| IR-09 | Migration mistakes on hot tables during expand-contract | L | H | scratch-env upgrade/downgrade test tier (test-matrix); per-service down migrations enforced | upgrade tier red | founder |
| IR-10 | Flake accumulation erodes trust in gates | M | M | quarantine+7-day SLA; >2% blocks train (already policy) | flake rate metric | founder |
