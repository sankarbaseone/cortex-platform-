# RB-DB-002 — NyduxChPartsHigh (page)
**Trigger:** >300 parts/partition — merge pressure, query cliff ahead. · **Owner:** on-call

## Symptoms
>300 parts/partition — merge pressure, query cliff ahead.

## Dashboards
NYDUX / Service Overview (+ domain dashboard per dashboard-registry).

## Diagnosis tree
1. Insert pattern change? async_insert disabled by config drift is the classic cause.
2. Merges stuck: disk headroom <20%? system.merges errors?
3. TTL moves to cold volume backed up? S3 throttling.

## Mitigation
Restore async_insert; free disk; OPTIMIZE TABLE ... FINAL only as last resort off-peak.

## Post-actions
Alert earlier next time? consider 200 warn threshold if repeat.
