# RB-BUS-004 — NyduxAuditGapDetected (Sev-1)
**Alert:** `increase(nydux_audit_chain_gaps_total[5m]) > 0` · **Severity:** page, Sev-1 · **Owner:** platform on-call

## Symptoms
Audit chain sequence gap detected for one or more tenants. This is a potential
tamper/loss event on the compliance ledger (RFC-009 I.8). Treat as security
incident until proven otherwise.

## Dashboards
- NYDUX / Service Overview → service=audit-svc (lag, DLQ, outbox)
- NYDUX / Audit Chain (per-tenant seq watermarks)

## Diagnosis tree
1. `nyduxctl audit verify --tenant <t> --from <last_good_seq>` — recompute chain.
2. Gap coincides with Kafka partition failover? → check broker events; if yes,
   likely delayed append, NOT loss: wait one poll cycle, re-verify.
3. Gap persists → compare PG `audit_entries` max(seq) vs Kafka high-watermark:
   - PG ahead of Kafka → outbox relay stalled → restart audit-svc relay; re-verify.
   - Kafka ahead of PG → consumer crash mid-batch → replay from last committed offset.
4. Hash mismatch (not just gap) → SECURITY INCIDENT: freeze writes
   (`nyduxctl tenant freeze <t>`), page security lead, preserve broker segments,
   compare against daily object-lock anchor (RFC-009 I.8) to bound tamper window.

## Mitigation
Gap-only: replay/redrive per step 3; verify returns valid; close.
Mismatch: incident process RB-SEC-001; customer notification clock (≤72h) starts
at confirmation (RFC-009 I.12).

## Post-actions
Post-incident review mandatory (Sev-1). Add regression test reproducing the
failure mode to audit-svc itest suite before closing.
