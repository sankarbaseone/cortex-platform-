# RB-SEC-001 — Security incident process
**Trigger:** Suspected compromise / audit tamper / data exposure. · **Owner:** on-call

## Symptoms
Suspected compromise / audit tamper / data exposure.

## Dashboards
NYDUX / Service Overview (+ domain dashboard per dashboard-registry).

## Diagnosis tree
1. Contain: freeze affected tenant writes (nyduxctl tenant freeze), rotate credentials in scope (ESO resync), preserve evidence (snapshots, broker segments, object-lock anchors).
2. Assess: scope via audit chain + access logs; classify data classes touched (D.9 categories).
3. Notify: customer clock <=72h from confirmation (RFC-009 I.12); regulator per DPA if EU tenant.
4. Eradicate+recover: patched images via emergency train (expedited J.6, gates NOT skipped).

## Mitigation
Founder is incident commander until security hire; external IR retainer contact in vault.

## Post-actions
Mandatory post-incident within 5 business days; threat-model row update.
