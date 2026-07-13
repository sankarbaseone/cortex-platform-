# RB-BUS-002 — NyduxDLQNonEmptyBusiness (page)
**Trigger:** Any message in a business-class DLQ (threshold 0). · **Owner:** on-call

## Symptoms
Any message in a business-class DLQ (threshold 0).

## Dashboards
NYDUX / Service Overview (+ domain dashboard per dashboard-registry).

## Diagnosis tree
1. `nyduxctl dlq inspect --topic T` — read envelope + poison header.
2. poison=true (unmarshal fail): producer shipped bad schema? check NyduxSchemaRegistryDrift; else consumer bug.
3. Retries-exhausted: read consumer error in DLQ headers; usually downstream store outage window.
4. Confirm idempotency class of event (topics.yaml) before redrive.

## Mitigation
Fix root cause, then `nyduxctl dlq redrive --topic T --dry-run` -> real. Never drop business events without founder sign-off (audited).

## Post-actions
Add the failing payload as consumer test fixture.
