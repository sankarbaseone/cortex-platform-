# ECD-011 — Observability Construction Documents
**Level:** 2 · Extends RFC-011 J.8/J.9, RFC-014 G.3, ECD-007 §7.5 (frozen). **Artifacts are the spec.**

## Shipped artifacts
- `artifacts/observability/prometheus/nydux-rules.yaml` — complete PrometheusRule CRD: SLO recording rules + 18 alerts across bus (ECD-007 fixed names, exact thresholds), SLO burn (99.9 API fast-burn math), data plane (clock skew, spool fill, privacy-reject, MEASUREMENT_ONLY spike → parser-matrix signal), stores (dead tuples, CH parts, Timescale jobs), agents (schema-reject spike = prompt drift, judge-fail loop). Every alert carries its runbook path.
- `artifacts/observability/slo/slos.yaml` — Sloth definitions for the three contractual SLOs (API 99.9, ingest 99.95, first-score freshness 95%-within-5min) generating multi-window burn alerts.
- `artifacts/observability/grafana/service-overview.json` — complete importable dashboard (7 panels, service variable): RPS by code, latency quantiles, consumer lag, DLQ depth, outbox backlog, dependency p95, breaker state.
- `artifacts/observability/runbooks/RB-BUS-004.md` — complete Sev-1 runbook (audit chain gap) as the NORMATIVE template; `runbook-registry.yaml` enumerates all 20 required runbooks with owners; CI rule: alert without registry entry fails; `body-required` entries block GA.
- `artifacts/registries/metrics-registry.yaml` — canonical closed metric vocabulary (uniform interceptor set + per-domain), with the cardinality rule (tenant label only on aggregated views).
- `artifacts/registries/alerts-registry.yaml`, `dashboard-registry.yaml` — synchronized indices; dashboard registry fixes the required panel set for the 6 remaining dashboards and the generation rule (layout template = shipped JSON).

## Construction rules
Tracing: OTel SDK everywhere, W3C propagation, 1% head sampling + tail-based on error (collector config in charts); span naming `svc.Method`. Logging: JSON schema fixed (ECD-002 §1), payload-content logging lint-forbidden. Error budgets: Sloth output wired to release gating (RFC-011 J.6/J.11) — burn >2%/h pauses the canary automatically (Argo Rollouts analysis template referencing `nydux:api_availability:ratio5m`).
