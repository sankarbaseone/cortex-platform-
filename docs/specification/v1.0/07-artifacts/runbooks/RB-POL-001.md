# RB-POL-001 — NyduxPolicyDecisionSlow (page)
**Trigger:** Policy p95 >50ms — in the approve/deploy path. · **Owner:** on-call

## Symptoms
Policy p95 >50ms — in the approve/deploy path.

## Dashboards
NYDUX / Service Overview (+ domain dashboard per dashboard-registry).

## Diagnosis tree
1. Bundle reload storm? (policy edits) — reload p95 panel.
2. Cache hit collapse: TTL misconfig or input cardinality explosion (agent tools flooding Decide).
3. OPA eval regression after policy change: profile with opa eval --profile on the fixture set.

## Mitigation
Rate-limit offending caller; revert policy change; scale policy-svc.

## Post-actions
Add the slow policy to the fixture perf suite.
