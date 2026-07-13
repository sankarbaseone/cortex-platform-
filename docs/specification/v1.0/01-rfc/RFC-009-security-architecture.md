# RFC-009 — Security Architecture

**Status:** Approved · **Extends:** V1 Phase 5 (RBAC/audit/secrets/tenancy), Phase 14 risk #6 · **Owns Section I** entirely.

## I.1 Compliance map
| Framework | Posture | Notes |
|---|---|---|
| SOC 2 Type II | target end Y2 (V1 roadmap) | controls implemented from Sprint 1; evidence automated (audit chain, CI attestations) |
| ISO 27001 | Y3 | shares control set |
| ISO 42001 (AI mgmt) | Y3 | agent governance (RFC-008) maps directly |
| NIST 800-53 mod / CSF | mapped now | FedRAMP delta = paperwork per OQ-08 |
| EU AI Act Art.12 | product feature | audit ledger + retention ≥6mo (we keep ≥7y, RFC-007 C.8) |
Control catalog lives in `compliance/controls.yaml` — each control lists implementing components + automated evidence query.

## I.2 Zero Trust
No implicit trust by network location. Every request: authenticated principal (human OIDC / workload SPIFFE mTLS), authorized per-call (RBAC+ABAC), encrypted in transit (TLS1.3 everywhere, mesh mTLS internal). DP↔CP: mutual TLS with cert-bound tokens; gateway pinning of CP CA. Admin access: no SSH to prod; break-glass via short-lived certs (30-min), dual-approved, session-recorded.

## I.3 Encryption & KMS
At rest: all stores encrypted; **per-tenant data keys** (envelope encryption, DEKs wrapped by tenant KEK in KMS; BYOK supported for dedicated/self-hosted). Key rotation: KEK yearly, DEK on demand + on offboarding (crypto-shredding = offboarding guarantee). In transit: TLS1.3 min. Field-level: API key hashes (SHA-256+pepper), audit entry hashes BLAKE3.

## I.4 Secrets
Vault (or cloud KMS+ESO) — no secrets in env-files/images/git (gitleaks in CI). Workload identity for cloud creds (no static keys). DP collectors need zero customer secrets (read-only host metrics + K8s SA with minimal Role: get/list pods,nodes).

## I.5 RBAC + ABAC
Roles (RFC-007 tables): `viewer, engineer, approver, finops, auditor, tenant-admin, nydux-operator(support, customer-grantable, time-boxed)`. ABAC attributes: team scope, cluster scope, environment (prod/staging), data-class. Enforcement points: API coarse role check → service-level OPA sidecar policy (`authz.rego`) with attributes → DB RLS as final backstop. Approver for risk>θ actions requires role `approver` AND scope match AND not-author (separation of duties).

## I.6 Multi-tenancy isolation (defense in depth)
(1) JWT tenant claim → request GUC → (2) RLS on every PG/Timescale table → (3) ClickHouse row policies → (4) per-tenant graph namespaces → (5) per-tenant KMS keys → (6) per-tenant blob buckets/prefixes with bucket policy → (7) egress privacy filter (RFC-003 D.9). Isolation test suite: automated cross-tenant probe attempts run nightly in staging; any success = Sev-1.

## I.7 Threat model (STRIDE summary, top items)
| Threat | Vector | Control |
|---|---|---|
| Tenant data exfil via global graph | over-broad features | egress allowlist + k-anonymity (D.9) + CI privacy test |
| Spoofed collector | stolen cert | SPIFFE identity + cert rotation 24h + cluster attestation nonce |
| Malicious IR upload (parser exploit) | crafted PTX/SASS | Rust parsers, fuzzing (RFC-002 2.9), analyzer sandbox (gVisor/no-net pod, seccomp) |
| Prompt injection → unauthorized action | retrieved data | RFC-008 H.6: schema-validated tools, approval tokens, no-write agents |
| Supply chain | dep compromise | I.9 |
| Insider (NYDUX) | support access | time-boxed grants, session recording, customer-visible access log |
| Audit tampering | DB write | hash chain + external anchor (daily chain-head published to object-lock bucket + optional customer webhook) |

## I.8 Audit
Append-only per-tenant chain: `entry_hash = BLAKE3(prev_hash || canonical(entry))`; single Kafka partition per tenant (RFC-005 B.5); continuous gap/chain verifier service; `/v1/audit/verify` recomputes on demand. What is audited: authn events, all writes, approvals, policy decisions, agent tool calls (args-hash), support access, exports, baseline re-anchors.

## I.9 Supply chain & SBOM
SLSA-3 target: hermetic builds (Bazel or locked containers), provenance attestations (in-toto), cosign-signed images + charts, SBOM (Syft SPDX) published per release, Renovate + osv-scanner gating, base images distroless, admission policy verifies signatures in customer clusters (optional but default-on in our charts).

## I.10 OWASP & runtime security
ASVS L2 checklist in PR template for API changes; dependency- and container-scanning in CI; runtime: Falco rules on DP/CP pods (exec-into-container, unexpected outbound); WAF on public API; per-tenant anomaly alerts on API usage.

## I.11 Security testing
SAST (semgrep custom rules incl. "no raw IR egress" pattern), DAST on staging weekly, external pentest 2×/yr, red-team exercise on agent framework 1×/yr (prompt-injection corpus maintained in repo), fuzzing continuously (OSS-Fuzz-style cluster for parsers), threat-model review required in design docs touching trust boundaries (RFC-011 template).

## I.12 Incident response
Sev matrix, 24×7 on-call (RFC-011 J.11), IR runbooks (`RB-SEC-*`), customer notification ≤72h contractual (GDPR-aligned), post-incident review public-to-customer for Sev-1/2.
