# RB-KES-001 — NyduxFirstScoreFreshnessSLO (page)
**Trigger:** First-score p95 >5min. · **Owner:** on-call

## Symptoms
First-score p95 >5min.

## Dashboards
NYDUX / Service Overview (+ domain dashboard per dashboard-registry).

## Diagnosis tree
1. Where is the queue? gateway spool vs kernel-registry consumer lag vs analyzer duration panel (mode=static vs ncu).
2. Analyzer slow: cache hit ratio dropped? (kes model bump invalidates cache — expected transient).
3. Registry slow: PG upsert latency / outbox contention.

## Mitigation
Scale the bottleneck stage; if model-bump transient, annotate and watch; freshness recovers as cache refills.

## Post-actions
If parser gap (see RB-KES-002) added latency, prioritize matrix entry.
