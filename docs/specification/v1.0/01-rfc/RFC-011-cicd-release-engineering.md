# RFC-011 — CI/CD & Release Engineering

**Status:** Approved · **Owns Sections J and Q.**

## J.1 Repository layout
Monorepo `nydux/platform` (atomic cross-service changes, one toolchain, shared protos). Separate repos: `nydux/sdk-python` (release cadence differs), `nydux/helm-charts` (customer-facing), `nydux/docs`.
```
platform/
  proto/            # single source of truth; buf managed
  services/<name>/  # Go services (see RFC-014)
  analyzers/        # Rust: collector, parsers, edge-gateway
  agents/           # Python: orchestrator, agents, prompts/
  web/              # React app + BFF
  libs/{go,rs,py,ts}/
  charts/ deploy/ infra/(terraform) compliance/ docs/ tools/
```

## J.2 Git & branch strategy
Trunk-based: `main` always releasable; short-lived branches `feat/…`, `fix/…`; squash-merge; Conventional Commits (feeds changelog + semver); no long-lived release branches except enterprise LTS (`lts/v1` cut at GA, cherry-pick fixes only, 36-month life per V1).

## J.3 CI pipeline (per PR; GitHub Actions or Buildkite)
lint(golangci, clippy, ruff, eslint) → typecheck → unit tests (race detector on) → buf breaking-change check → schemathesis API fuzz (changed endpoints) → build images (hermetic, cached) → integration tests (kind cluster + testcontainers: PG/CH/Kafka/Redis) → e2e smoke (ephemeral env, 15-min budget) → security (semgrep, gitleaks, osv, container scan) → coverage gate → benchmark diff (hot paths only, ±5% budget) → SBOM + sign. Target wall time <20 min via test-impact analysis.

## J.4 Testing pyramid (ratios enforced by convention, reviewed quarterly)
- Unit (~70%): pure logic; KES math, CRI stats, savings properties (RFC-004 4.11), parsers (plus fuzz corpus).
- Integration (~20%): service+store; Kafka consumer idempotency tests mandatory per consumer.
- E2E (~8%): golden path — install chart → synthetic kernel fixture → score visible → rec created → approve → verify. Runs on every main merge + nightly full matrix.
- Non-functional (~2%): load (k6: API; synthetic 10k-GPU telemetry generator for ingest), chaos (Litmus: broker kill, AZ loss, analyzer OOM — RFC-005 B.9, RFC-001 A.14 scenarios), soak weekly 24h.

## J.5 Migrations
golang-migrate; PR must include: forward migration, rollback note, `--dry-run` output attached, backfill plan if >1M rows (batched, resumable, off-peak). Expand-contract pattern mandatory for renames (never break running old version during deploy).

## J.6 Release strategy
Weekly train from `main` → staging (auto) → canary in prod SaaS: 1 tenant-cohort (internal + opted-in design partners) → 10% → 100%, gated on SLO burn + error-budget + key business probes (kernel-scoring freshness, ingest lag). **Blue-green** for stateful-adjacent services (API tier); **canary** for domain services; DP components (collector/analyzer) versioned independently, customer-controlled upgrade windows, N-2 compatibility guaranteed and tested in matrix CI. Rollback = redeploy previous digest (≤5 min) + migration rollback note if applicable; feature flags (OpenFeature + flagd) for anything user-visible — kill without deploy.

## J.7 Artifact & environment promotion
Images immutable by digest; the same digest promotes dev→staging→prod (no rebuilds); provenance attestation verified at admission (RFC-009 I.9). Envs: `dev` (ephemeral per-PR, vcluster), `staging` (prod-shaped, synthetic tenants), `prod` regions, `airgap-sim` (quarterly bundle rehearsal, RFC-001 A.13).

## J.8 Observability of NYDUX itself
OTel everywhere (traces 1% sampled, tail-based on errors); RED metrics per service + USE per node; SLOs codified (Sloth) with burn alerts; logs structured JSON, PII-free by lint rule; dashboards as code (Grafana provisioning) shipped with each service (RFC-014 requirement).

## J.9 Runbooks
`docs/runbooks/RB-<AREA>-<NNN>.md`; required sections: symptoms, dashboards, diagnosis tree, mitigation, escalation, post-actions. Every alert links a runbook (CI check: alert without runbook fails).

## J.10 Continuous verification
Monthly automated DR restore drill (RFC-001 A.7) with report; nightly cross-tenant isolation probes (RFC-009 I.6); weekly forecast backtests (RFC-004 4.11); quarterly chaos game-day.

## J.11 On-call
Two rotations (platform, data/ML) follow-the-sun as team grows; Sev1 page, Sev2 ticket-with-page-hours; error budgets: feature freeze when 30-day budget spent.

## Q — Quality Gates (every PR; enforced, not advisory)
| Gate | Rule | Tooling |
|---|---|---|
| Architecture | import/dependency rules pass (layering RFC-001 A.2) | depguard/import-linter |
| Security | 0 high/critical findings; secrets scan clean; trust-boundary change ⇒ threat-model section in PR | semgrep, gitleaks, osv |
| Performance | hot-path benchmarks within ±5%; new endpoint has k6 smoke | benchdiff |
| Code quality | lint clean; cyclomatic hotspots reviewed; no TODO without ticket | linters |
| Docs | public API change ⇒ OpenAPI/proto docs updated; user-visible ⇒ docs PR linked | CI check |
| Testing | new logic has unit tests; consumer changes include idempotency test | review + coverage |
| Coverage | changed-lines coverage ≥80% (not global vanity) | diff-cover |
| Observability | new service/endpoint ships metrics+dashboard+alert+runbook | checklist bot |
| Privacy | egress schema unchanged or privacy-review label present (D.9) | CI schema diff |
| Approvals | 1 reviewer; 2 for trust-boundary/migration/proto-breaking | CODEOWNERS |
