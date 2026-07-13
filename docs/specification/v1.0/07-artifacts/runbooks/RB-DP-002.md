# RB-DP-002 — NyduxGatewaySpoolFilling (page)
**Trigger:** Spool >80% — CP link degraded; data loss at 100%+24h. · **Owner:** on-call

## Symptoms
Spool >80% — CP link degraded; data loss at 100%+24h.

## Dashboards
NYDUX / Service Overview (+ domain dashboard per dashboard-registry).

## Diagnosis tree
1. Egress reachable from DP? (customer firewall change is the usual cause).
2. Our ingest 5xx? correlate with NyduxApiAvailabilityBurnFast / ingest SLO.
3. Single cluster vs many: many = our side.

## Mitigation
Our side: scale ingest path, fix, spool drains with replay=true (watermark absorbs). Customer side: support escalation with spool ETA to overflow (panel).

## Post-actions
If overflow occurred: quantify loss window from spool_id gaps, notify per contract.
