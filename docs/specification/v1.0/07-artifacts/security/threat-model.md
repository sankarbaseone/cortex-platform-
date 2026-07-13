# NYDUX V1.0 Threat Model (ECD-012 §1) — STRIDE × trust boundaries
Trust boundaries: TB1 customer cluster↔edge-gateway (mTLS/SPIFFE) · TB2 DP↔CP (gRPC egress-only) · TB3 CP internal mesh · TB4 user/browser↔API · TB5 CP↔LLM provider · TB6 CP↔cloud KMS/S3.

| ID | Boundary | STRIDE | Threat | Mitigation (implemented artifact) | Residual |
|---|---|---|---|---|---|
| T-01 | TB1 | S | Rogue collector impersonates cluster | single-use bootstrap token→SPIFFE cert (ECD-005 5.1); cert TTL 24h | low |
| T-02 | TB2 | I | Sensitive IR exfil via telemetry | D.9 allowlist filter at gateway, reject=hard error + NyduxEgressPrivacyReject alert; semgrep no-raw-IR rule in CI | low |
| T-03 | TB2 | T | Event tampering in transit | mTLS + envelope schema-validate + BLAKE3 kernel identity end-to-end | low |
| T-04 | TB3 | E | Lateral movement CP pod→DB | default-deny NetworkPolicy (chart), per-service DB creds (ESO), RLS FORCE | low |
| T-05 | TB4 | S | Session/API-key theft | 15-min JWT + refresh rotation; api-key hash-at-rest, prefix lookup, scope ci-only | med (customer IdP hygiene) |
| T-06 | TB4 | E | Author self-approves rec | SoD enforced server-side (F-AT-01 verify not-author) + UI hide + Playwright test | low |
| T-07 | TB5 | I | Prompt injection via tenant data in agent context | tool schema server-side validation, grounding=1.0 judge, no write without human token (RFC-008 H.5/H.10); injection eval set in agent CI | med — accepted, mitigated by human gate |
| T-08 | TB5 | D | Agent cost runaway/loops | per-task budget_usd_micros hard stop, kill-switch flag agents.enabled, judge-fail-loop alert | low |
| T-09 | TB3 | R | Operator denies destructive action | hash-chained audit (F-AU-01), daily object-lock anchor, gap detector Sev-1 | low |
| T-10 | TB3 | T | Audit chain tamper by DB admin | append-only grants, chain verify vs object-lock anchor, RB-BUS-004 | low |
| T-11 | TB6 | I | Backup exfil | KMS per-tenant DEK envelope encryption; bucket policy deny non-VPC | low |
| T-12 | TB4 | D | API abuse/scraping | per-tenant rate limits (registry), WAF at LB, 429+RateLimit headers | low |
| T-13 | supply | T | Malicious dependency/image | SBOM+osv-scanner+cosign signing+digest pinning; SLSA-3 provenance in CI | med (0-day window) |
| T-14 | TB1 | D | Spool exhaustion attack on gateway | spool cap + backpressure to collector + NyduxGatewaySpoolFilling | low |
| T-15 | TB4 | I | Cross-tenant read via filter injection | RSQL closed whitelist parse (never string-concat SQL), RLS as backstop, itest cross-tenant-fails | low |

Abuse cases tested in CI (security suite): self-approval, expired/other-user approval token, replayed apply, RSQL injection strings, cross-tenant UUID probing, agent tool-arg schema violation, oversized batch, privacy-field smuggling in map keys.
