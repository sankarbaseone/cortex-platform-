# RB-AG-002 — NyduxAgentJudgeFailLoop (ticket)
**Trigger:** >30% judge failures — agents burning budget without output. · **Owner:** on-call

## Symptoms
>30% judge failures — agents burning budget without output.

## Dashboards
NYDUX / Service Overview (+ domain dashboard per dashboard-registry).

## Diagnosis tree
1. Which kind? grounding failures (citations missing) vs verdict failures.
2. Provider degradation (latency/5xx) causing truncated outputs?
3. Knowledge staleness: graph/MV freshness watermark old?

## Mitigation
Pause the kind via flag if budget waste material; fix upstream freshness; provider failover per config.

## Post-actions
Eval-set addition; budget guard verified (no task exceeded cap).
