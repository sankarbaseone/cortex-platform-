# RB-DB-003 — NyduxTimescaleJobFailed (ticket)
**Trigger:** Compression/retention/cagg job errors. · **Owner:** on-call

## Symptoms
Compression/retention/cagg job errors.

## Dashboards
NYDUX / Service Overview (+ domain dashboard per dashboard-registry).

## Diagnosis tree
1. timescaledb_information.job_errors: which job + sqlstate.
2. Lock conflict with long analytics query is most common (cagg refresh).
3. Disk full on compress: check chunk sizes.

## Mitigation
Re-run job off-peak; kill conflicting query; extend maintenance window.

## Post-actions
Recurring lock conflicts: move cagg schedule.
