# RB-BUS-006 — NyduxSchemaRegistryDrift (ticket)
**Trigger:** Producer emitting schema id unknown to registry. · **Owner:** on-call

## Symptoms
Producer emitting schema id unknown to registry.

## Dashboards
NYDUX / Service Overview (+ domain dashboard per dashboard-registry).

## Diagnosis tree
1. Identify producer from metric label; compare its image digest vs release train.
2. Was buf-breaking green? If yes: registry sync job failed — rerun tools/genmanifests publish.
3. If producer is canary: expected during dual-publish? verify .v2 topic plan.

## Mitigation
Sync registry or roll back producer canary.

## Post-actions
Post-incident: registry publish becomes a release-gate step if it wasn't the cause.
