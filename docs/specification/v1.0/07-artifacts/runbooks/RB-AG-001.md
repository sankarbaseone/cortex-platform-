# RB-AG-001 — NyduxAgentSchemaRejectSpike (ticket)
**Trigger:** Tool-arg schema rejects >10/15m — prompt drift or injection probing. · **Owner:** on-call

## Symptoms
Tool-arg schema rejects >10/15m — prompt drift or injection probing.

## Dashboards
NYDUX / Service Overview (+ domain dashboard per dashboard-registry).

## Diagnosis tree
1. Which tool + task kind? Sample rejected args (audited hashes -> transcripts).
2. After model/prompt version bump? diff prompt registry version.
3. Same tenant, adversarial-looking args: treat as T-07 probing — review transcripts.

## Mitigation
Roll back prompt version; if probing: tenant security contact + tighten input source.

## Post-actions
Add rejected patterns to injection eval set.
