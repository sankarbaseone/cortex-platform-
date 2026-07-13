# RB-DR-001 — DR region promote (manual)
**Trigger:** Primary region loss — promote pilot-light DR (RFC-001 A.7). · **Owner:** on-call

## Symptoms
Primary region loss — promote pilot-light DR (RFC-001 A.7).

## Dashboards
NYDUX / Service Overview (+ domain dashboard per dashboard-registry).

## Diagnosis tree
1. Declare: founder/on-call decides; set api.readonly=true globally if API partially up.
2. Promote PG/TS replicas in DR (walg latest+WAL); restore CH from S3 backup + Kafka replay window.
3. Flip DNS to DR gateway; scale DR node groups from pilot to prod sizes (terraform apply -var profile=dr-active).
4. Verify: audit chain continuity per tenant (Verify from last anchor); ingest resumes (DP spools drain).

## Mitigation
RTO target 1h; publish status page timeline; RPO<=5min verified via WAL timestamps.

## Post-actions
Full post-incident: measure achieved RTO/RPO vs target, drill delta list.
