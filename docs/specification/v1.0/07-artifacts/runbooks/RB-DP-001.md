# RB-DP-001 — NyduxCollectorClockSkew (ticket)
**Trigger:** DP node clock skew >2s — telemetry ordering risk. · **Owner:** on-call

## Symptoms
DP node clock skew >2s — telemetry ordering risk.

## Dashboards
NYDUX / Service Overview (+ domain dashboard per dashboard-registry).

## Diagnosis tree
1. Confirm chrony/NTP health on the node (customer-side).
2. Widespread on one cluster: customer NTP source issue — open support thread with evidence.
3. Check late-row ratio: is analytics watermark absorbing it?

## Mitigation
Customer remediation guide (docs/ops/ntp.md); our side: nothing to change — watermark handles ≤6h.

## Post-actions
Track recurrence per cluster; persistent skew -> onboarding checklist addition.
