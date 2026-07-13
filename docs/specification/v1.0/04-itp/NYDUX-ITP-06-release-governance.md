# Release Governance — stage gates

| Stage | Definition | Mandatory gates |
|---|---|---|
| **Alpha** (internal) | M2 vertical slice on staging; NYDUX-internal tenant only | CI Q gates · M2 exit criteria · seeded-corpus first-score SLO met · no Sev-1 open |
| **Beta** (design partners, ≤3 tenants) | M4 value loop + M3 gates usable | Alpha + full rec-lifecycle e2e · audit verify green in staging for 14 days · RLS/SoD security tests green · onboarding TTFHW ≤1 day proven with partner #1 · support channel + RB-SEC-001 active |
| **Release Candidate** | M6 feature-complete, hardening in progress | Beta + full Playwright+axe · load test budgets green · chaos CH-01..08 expectations met · 72h soak clean · all runbook bodies complete (registry gate) · error budget >50% remaining over trailing 30d |
| **GA** | enterprise-sellable | RC + DR drill restore <1h with audit continuity · external pentest criticals=0/highs remediated · monthly drill calendar active · SOC2 evidence collection running · savings replay bundle produced for a real partner month · founder sign-off recorded in audit chain |

Regression rule between stages: any Sev-1, chain gap, or budget regression >10% demotes the stage until cleared. Stage status lives on the Founder Dashboard.
