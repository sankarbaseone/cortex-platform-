# Engineering Delivery Plan (sequencing only — features are frozen)

## Milestones
| M | Name | Meaning | Exit criteria |
|---|---|---|---|
| M0 | Green Skeleton | repo + CI + libs + schemas live | S0 done; empty-green pipeline; RLS itest red-green proven |
| M1 | Ingest Spine | telemetry flows device→ClickHouse | S1 done; e2e fixture queryable; privacy filter proven |
| M2 | **Vertical Slice** (critical demo) | one real kernel: submit→canonicalize→KES→registry→UI-visible score | T-201..204 + minimal kernels screen; first-score <5min on corpus |
| M3 | Regression & Gates | CRI + CI gate usable by a design partner | S2 done; `nydux regressions --fail-on` exits correctly against staging |
| M4 | Value Loop | rec lifecycle end-to-end with policy+audit | S3 done; approve→apply→verify→rollback e2e; audit verify green |
| M5 | Economics | cost attribution + twin + savings dual-report | S4 done; sum-preservation + MAPE gates green |
| M6 | Full Surface | all 13 screens + agents + notify | S5 done; full Playwright + axe green |
| M7 | Hardened GA | chaos/soak/load/DR/pentest passed | S6 done; ITP-06 GA gates all green |

## Epics = build-order sprints (S0→S6); stories = tasks T-001→T-603 (29). No re-sequencing permitted except via founder note in progress.yaml.

## Critical path
T-001→T-002→T-003→T-101→(T-201→T-202/T-203→T-204)→T-205→T-301→T-302/T-303→T-304→T-501→T-601→T-603. Everything else hangs off it.

## Parallel opportunities (safe lanes for multi-instance Claude Code)
Lane A (Go spine): T-101→T-102→T-205… · Lane B (Rust): T-201→T-202/T-203 independent until T-204 joins · Lane C (platform): T-104, T-103 after T-002 · Lane D (surface): T-501 component library can start against generated api-client mocks after T-002; feature screens gate on their APIs. Rule: lanes never share a service directory; proto/ and registries/ changes serialize through the contract-review rule.

## Acceptance criteria
Per task: its DoD in build-order.yaml. Per milestone: the exit criteria above. Per release stage: ITP-06.

## Sprint goals (nominal 2-week cadence, solo+Claude Code velocity assumption; recalibrate after S0 actuals)
S0 skeleton · S1 ingest · S2 compiler intelligence (2 sprints likely — Rust crates + analyzer) · S3 value loop · S4 economics · S5 surface (2 sprints likely) · S6 hardening. Nominal 8–9 sprints ≈ 16–18 weeks to GA-ready; the 90-day plan (ITP-09) targets M0–M5 inside the first 13 weeks.
