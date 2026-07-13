# ECD-009 — Frontend Construction Documents
**Level:** 2 · Extends RFC-010 (frozen). **Artifacts are the spec.**

## Shipped artifacts
- `artifacts/frontend/routes/router.tsx` — complete TanStack route table: 21 routes covering all 13 screens + detail routes; lazy feature bundles.
- `artifacts/frontend/component-tree.yaml` — canonical component inventory (validated): 11 shared components with props contracts, per-feature composition for every screen, loading/error/empty state policy, keyboard + a11y policy. New components require a PR touching this file (drift rule).
- `artifacts/frontend/playwright/core-flows.spec.ts` — executable e2e suite for the primary flows: onboarding live-connect flip, kernel explorer filter→detail→SaaS IR privacy explainer, full rec lifecycle (rationale required, SoD author-cannot-approve, apply→rollback visible), CI-gate snippet clipboard exactness, savings re-anchor dual co-sign with audit link, axe AA gate.

## Construction rules
- Data layer: generated `libs/ts/api-client` from openapi.yaml only — hand-written fetches forbidden (lint rule `no-restricted-imports` on fetch/axios outside api-client).
- Query keys: `[operationId, params]` exactly; SSE channels: `/dashboard` only (OQ-15); WS only graph.
- Test ids: `data-testid` names used in the Playwright suite are NORMATIVE — components must ship them (kes-radar, kes-components, rec-state, rollback-button, sod-explainer, helm-command, cluster-status-*, cri-heatmap, baseline-version, audit-link, ir-privacy-explainer, freshness watermarks).
- State: TanStack Query for server state; session context only client state; no Redux/Zustand (rejected: server-cache duplication).
- Charts: visx components per tree; every chart has a table alternative (AA).
- Error mapping: ProblemBanner maps problem+json `type` codes (ECD-003 §3.7 catalog) to UX copy table in `web/app/src/lib/problems.ts` (one entry per catalog code — closed set, compile-checked).

## Remaining gap (declared, not silent)
Static wireframe images are NOT generated (an image set adds no determinism beyond component-tree.yaml + testids + Playwright assertions, which together fully constrain layout semantics). If pixel-level mocks are wanted, that is a design task, not an engineering ambiguity.
