# RB-BUS-003 — NyduxDLQTelemetryBurst (ticket)
**Trigger:** Telemetry DLQ >1000 — data-quality, not availability. · **Owner:** on-call

## Symptoms
Telemetry DLQ >1000 — data-quality, not availability.

## Dashboards
NYDUX / Service Overview (+ domain dashboard per dashboard-registry).

## Diagnosis tree
1. Identify reject reason distribution from DLQ headers (validate_failed vs too_old vs schema).
2. too_old burst: check collector clock skew alert; validate burst after collector release: version skew — check dual-publish window.
3. Sustained burst from one cluster: that cluster's collector config drifted.

## Mitigation
Fix source; telemetry DLQ may be expired (14d) rather than redriven if data superseded by rollups.

## Post-actions
Note in weekly ops review; adjust validation or collector pin.
