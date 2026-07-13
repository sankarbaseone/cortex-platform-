# RB-BUS-001 — NyduxConsumerLagHigh (page)
**Trigger:** Consumer lag beyond class threshold (business>60s / telemetry>300s). · **Owner:** on-call

## Symptoms
Consumer lag beyond class threshold (business>60s / telemetry>300s).

## Dashboards
NYDUX / Service Overview (+ domain dashboard per dashboard-registry).

## Diagnosis tree
1. Check Service Overview lag panel for topic+group; identify owning service (services.yaml consumers).
2. Consumer crash-looping? kubectl rollout status; if CrashLoop -> logs; poison message? check .dlq for same topic.
3. Throughput-bound? CPU throttling or partition skew (one partition hot) -> scale replicas toward partition count (HPA max in services.yaml).
4. Broker-side? under-replicated partitions / ISR shrink -> broker health first.
5. Backlog burn: after fix, lag should halve every ~5min; if flat, pause noncritical consumers of same broker.

## Mitigation
Scale consumers / restart stuck pod / redrive after fix. Business-class lag with user-visible staleness: post status page note.

## Post-actions
If poison message caused it: fixture + regression test on the consumer before close.
