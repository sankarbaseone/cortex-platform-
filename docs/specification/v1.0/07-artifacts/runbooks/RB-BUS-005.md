# RB-BUS-005 — NyduxOutboxBacklog (page)
**Trigger:** Unpublished outbox rows >10k for 5m — relay stalled. · **Owner:** on-call

## Symptoms
Unpublished outbox rows >10k for 5m — relay stalled.

## Dashboards
NYDUX / Service Overview (+ domain dashboard per dashboard-registry).

## Diagnosis tree
1. Which service (label)? Check its relay goroutine logs (publish errors).
2. Broker reachable? auth (SCRAM rotation!) is the common cause after secret rotation.
3. Row-level poison (oversized envelope)? relay skips+alerts — check size histogram.

## Mitigation
Restore broker path / rotate creds via ESO resync; relay drains automatically (batch 500/250ms ≈ 120k rows/min).

## Post-actions
If secret rotation caused it: add pre-rotation canary publish to rotation runbook.
