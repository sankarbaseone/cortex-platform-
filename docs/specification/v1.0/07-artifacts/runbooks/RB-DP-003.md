# RB-DP-003 — NyduxEgressPrivacyReject (page)
**Trigger:** Privacy filter rejecting egress — WORKING AS DESIGNED but signals producer bug. · **Owner:** on-call

## Symptoms
Privacy filter rejecting egress — WORKING AS DESIGNED but signals producer bug.

## Dashboards
NYDUX / Service Overview (+ domain dashboard per dashboard-registry).

## Diagnosis tree
1. Which field (metric label)? A DP component tried to ship a non-allowlisted field.
2. Recent DP release? diff the analyzer version — a new feature leaked a raw field.
3. If field SHOULD be allowed: that is a D.9 allowlist change = governance PR + privacy review, never a filter bypass.

## Mitigation
Roll back offending DP component. The filter stays on. No exceptions.

## Post-actions
Sev review: how did CI semgrep+fixture miss it; add fixture.
