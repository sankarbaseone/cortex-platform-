# RFC-010 — Frontend & UX + Product Requirements

**Status:** Approved · **Extends:** V1 Phase 4.10 (Control Plane) · **Owns Sections K and L.**

## K.1 Stack & rules
React 18 + TypeScript strict, Vite, TanStack Query/Router, visx for charts, Cytoscape.js for graph, WebGL fallback for >5k-node graphs. Design system: tokens-first (`@nydux/ui` package), dark-mode default (ops audience), density toggle. State: server-state via Query only; no global client store except session. Data: BFF GraphQL (read) + REST (mutations) — OQ-10. Live: SSE per-dashboard channel. i18n scaffolded (en only at GA). Accessibility: WCAG 2.1 AA — keyboard-complete, charts have table alternatives, contrast checked in CI (axe).

## K.2 Information architecture
```
Home(Dashboard) · Kernels(Explorer) · Compilers(Toolchains/Regressions) · Graph(Explorer)
Cost(Explorer) · Savings · Simulate(Twin) · Recommendations(Inbox) · Governance(Policies/Audit)
Agents(Advisor chat/Tasks) · Reports · Admin(Org/RBAC/API keys/Clusters/Settings) · Notifications
```

## K.3 Screens (each: purpose · primary components · key interactions · empty/error states)
1. **Dashboard** — fleet KES distribution, CRI trend, top-10 waste kernels ($), utilization vs cost sparkline, savings-to-date, freshness watermark. Click-through everywhere. Empty state = onboarding checklist (install collector → first kernels in <30min goal).
2. **Kernel Explorer** — virtualized table (hash, family, arch, KES, $/period, status badge), filter bar (RSQL builder), detail drawer: KES component radar, roofline plot with kernel dot, IR stage tabs (in-tenant deployments render IR; SaaS shows features-only with explainer — deliberate, D.9), similar-kernels panel, recommendations panel.
3. **Compiler Explorer** — toolchain registry with approval states (governance actions gated by role), regression matrix (from×to versions heat-map of CRI), pre-upgrade risk report (Q_REG_BLAST), CI-gate setup snippet.
4. **Knowledge Graph Explorer** — Cypher-free guided traversal (pick entity → expand relations), saved views, path explainer ("this $12k/mo traces to these 3 kernels"), WS transport.
5. **Cost Explorer** — dims: team/model/workload/kernel/token; stacked time-series + tree-map; unit-economics view (cost/1M tokens trend); export CSV.
6. **Savings** — verified savings per period, method dual-report (twin vs trailing), baseline versions timeline, re-anchor flow (dual co-sign wizard — legal text, both signatures, audit link).
7. **Simulation** — scenario builder (base snapshot picker + mutation forms: GPU SKU, parallelism, toolchain, quantization), Pareto chart, compare table, LOW_SUPPORT badges, "promote to capacity plan".
8. **Recommendations Inbox** — ranked list (score components visible), detail: evidence viewer (counters, IR spans), expected gain distribution, risk score, approve/reject with rationale (required), applied-state timeline with verify results and rollback button.
9. **Governance** — policy editor (Rego with lint + test-fixtures runner), decision log, audit chain browser + verify button, toolchain approval queue, EU-AI-Act evidence export wizard.
10. **Agents** — Advisor chat (citations rendered as chips linking to evidence), task list with transcripts (tool-call timeline), kill-switch (admin).
11. **Reports** — scheduled PDF/CSV (exec monthly savings, CRI report), template gallery.
12. **Admin/Org/RBAC** — users, roles matrix editor, scopes, API keys (create shows secret once), cluster registration (helm command generator with pre-filled token), settings (retention sliders within plan bounds, agent enablement, notification routes).
13. **Notification Center** — in-app + routes (email/Slack/webhook), per-event-type subscription matrix.

## K.4 Core UX flows (sequence)
**Onboard:** signup → create cluster token → copy helm cmd → collector heartbeat detected (live check) → first kernel scored → "aha" screen (top waste kernel + $). Target TTFHW (time-to-first-hard-win) <1 day.
**Approve rec:** inbox → detail → (optional) run verification benchmark → approve → watch applied/verify states → savings ledger updated.
**Pre-upgrade check:** Compiler Explorer → select from/to → risk report → add CI gate snippet.

## L — Product requirements (per feature bundle; format: story · acceptance · NFR)
**L.1 Kernel scoring**
- Story: As a perf engineer I see every executed kernel with KES and $ so I know what to fix first.
- Acceptance: kernel visible ≤5 min after first execution (measured path); KES components sum-explain score; static-only clearly badged; filter/sort p95 <500ms at 1M kernels.
- NFR: analyzer overhead per node ≤1% CPU; no raw IR egress in SaaS (automated test).
**L.2 Regression gate**
- Story: As a platform eng my CI fails when a toolchain bump risks >10% throughput.
- Acceptance: `nydux regressions --fail-on CRI>0.10` exits 2 with JUnit report; false-positive rate <10% on golden corpus; gate latency ≤5 min at 1k-kernel fleet (static pre-screen path).
**L.3 Savings ledger**
- Acceptance: every savings number traces to baseline version + method + replay bundle; negative savings displayed; re-anchor requires dual co-sign; auditor role read-only access.
**L.4 Recommendations**
- Acceptance: every rec has evidence refs, gain P50/P90, risk; apply requires approval token; rollback ≤1 click; verified gain writes RESULTED_IN edge.
**L.5 Governance**
- Acceptance: unapproved toolchain in prod (policy=block) prevents deploy event and logs decision; audit verify returns chain-valid in ≤10s for 1M entries.
**Global NFRs:** UI p95 route load <2s; API availability 99.9%; all screens AA; audit-logged mutations 100%; RTO/RPO per RFC-001 A.7; horizontal scale to 100k GPUs/tenant.
