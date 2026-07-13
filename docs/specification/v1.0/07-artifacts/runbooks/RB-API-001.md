# RB-API-001 — NyduxApiAvailabilityBurnFast (page)
**Trigger:** API 99.9 SLO fast burn (14.4x). · **Owner:** on-call

## Symptoms
API 99.9 SLO fast burn (14.4x).

## Dashboards
NYDUX / Service Overview (+ domain dashboard per dashboard-registry).

## Diagnosis tree
1. Scope: all routes or one? (RPS-by-code panel, group by route).
2. One route: owning service (api-registry) — check its deps panel + breaker state.
3. All routes: gateway/mesh/DB shared layer. Check control-plane-api and PG health first.
4. Deploy correlation: was a canary in flight? Argo analysis should have paused — check rollout state.

## Mitigation
Rollback in-flight release (make rollback SVC=x); if DB: failover per RB-DB path; enable api.readonly only for write-amplified incidents.

## Post-actions
Error-budget review; if canary gate missed it, tighten analysis query.
