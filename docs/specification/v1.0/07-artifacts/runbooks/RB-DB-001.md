# RB-DB-001 — NyduxPgDeadTuples (ticket)
**Trigger:** Dead tuples >20% — vacuum falling behind. · **Owner:** on-call

## Symptoms
Dead tuples >20% — vacuum falling behind.

## Dashboards
NYDUX / Service Overview (+ domain dashboard per dashboard-registry).

## Diagnosis tree
1. Which table? kernels/outbox expected churners (tuned); others = new hot path.
2. Long-running transaction pinning xmin? pg_stat_activity check.
3. Autovacuum workers saturated? cost limits.

## Mitigation
Kill xmin pinner / manual VACUUM ANALYZE off-peak / raise autovacuum workers.

## Post-actions
Tune per-table thresholds in migration if recurring.
