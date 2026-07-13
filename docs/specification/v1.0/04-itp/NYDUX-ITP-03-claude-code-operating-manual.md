# Claude Code Operating Manual (definitive)

## 0. Identity
You are the implementation engine. Authority: ITP-02 §1. Your constitution at repo root: CLAUDE.md (shipped). This manual is the operational expansion; CLAUDE.md wins on any perceived conflict.

## 1. Startup sequence (every session)
(1) Read CLAUDE.md → (2) read `registries/build-order.yaml`, locate the first task not marked done in `docs/progress.yaml` → (3) read that task's referenced ECD sections and exemplar → (4) confirm `make verify` is green on main before starting → (5) implement.

## 2. Repository initialization (Day 1 = task T-001)
Create the ECD-001 tree exactly; copy ALL shipped artifacts to their repo locations (protos → proto/, migrations → services/*/migrations, chart → charts/nydux-service, CI → .github/workflows, registries → registries/, runbooks → docs/runbooks, dashboards → observability/, Rego → services/policy-svc/policies, CLAUDE.md+README → root); emit `docs/spec-manifest.sha256`; initialize `docs/progress.yaml` (task_id → status/date/commit). Verify: `make bootstrap && make lint` green on the empty skeleton.

## 3. Development workflow (per task)
Branch `task/T-NNN-slug` → tests first per test-matrix tier requirements → implement to the F-spec → `make gen && make lint && make test` locally → `make itest` if adapters touched → update same-PR registry files if the task adds metrics/alerts/flags/endpoints → conventional commit `feat(scope): summary per ECD-XXX §Y [T-NNN]` → PR with DoD checklist → merge on green → mark task done in progress.yaml (same PR).

## 4. Module implementation workflow
Copy the exemplar service shape verbatim (domain/ports/adapters/cmd); implement F-specs in ECD-004 order for that service; wire main.go last; never leave a module with a public function lacking its required test tiers.

## 5. Verification workflow
Task-level: the task's `verify` command in build-order.yaml is the contract — it must pass, not approximately pass. Sprint-level: sprint exit = all tasks done + e2e smoke green + no open Sev-1/2 + progress.yaml updated. Budgets: any bench in the task scope asserts its ECD-014 number.

## 6. Commit & branch strategy
Trunk-based; short-lived task branches (<3 days — split the task if longer); one task per PR; squash-merge; no WIP on main; revert-first on red.

## 7. Documentation update policy
Code-reflecting docs only (godoc, generated API docs, runbook diagnosis details learned in incidents). NEVER touch RFC/ECD content — spec changes are ADRs. progress.yaml and the Founder Dashboard data file are updated every merge.

## 8. Testing expectations
Per test-matrix.yaml, non-negotiable: diff-cov ≥80%, required named tests present, property tests on math packages, golden byte-stability where specified, failure tests for every error path in the F-spec.

## 9. Release workflow
Release train per RFC-011: tag from main → release.yaml builds/signs/SBOMs → staging auto-deploy → canary 5→25→100 gated by Sloth burn → progress + dashboard updated → GA stages per ITP-06 gates.

## 10. Stop conditions (mandatory halt + founder approval)
Ambiguity (GAP protocol) · any change that would touch a frozen contract · a failing budget that resists optimization (candidate ADR) · security finding ≥ HIGH · two consecutive failed attempts at the same task's verify command (report, don't thrash).
