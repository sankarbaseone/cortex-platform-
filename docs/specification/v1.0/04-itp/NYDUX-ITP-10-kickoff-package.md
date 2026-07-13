# Implementation Kickoff Package — Day 1 instructions for Claude Code
*Permanent onboarding guide. Hand this file + repo access. Nothing else is needed.*

**You are building NYDUX V1.0 from Specification v1.0 (frozen). You have implementation authority only (ITP-02 §1).**

## Where to begin
1. Read `CLAUDE.md` at repo root (your constitution). Then ITP-03 (operating manual).
2. Open `registries/build-order.yaml`. Your first task is **T-001 repo-init**. Execute tasks strictly in order; parallel lanes only per ITP-04 and only when instructed by the founder.

## First module
T-001→T-005 (Sprint S0) produce the green skeleton — copy shipped artifacts verbatim, wire CI, prove RLS isolation. The first FEATURE module is **T-101 kernel-registry**, built by instantiating the exemplar (`artifacts/go-examples/`) — it is the reference shape for all 16 services.

## Which specifications to reference (per task)
The task entry names its files; its ECD sections are authoritative for HOW; F-specs (ECD-004) for function contracts; registries for every NAME. If a name isn't in a registry or ECD, it does not exist — do not invent it.

## How progress is verified
A task is complete ONLY when its `verify` command passes and its DoD holds, plus the module-level DoD checklist (ITP-05) in the PR. Update `docs/progress.yaml` in the same PR; `make dashboard` regenerates the founder view.

## When to stop and request approval
Mandatory halts (ITP-03 §10): genuine ambiguity → GAP protocol; anything touching a frozen contract → ADR proposal; budget that resists optimization; security finding ≥HIGH; two consecutive failures of the same verify command. Stopping correctly is success behavior, not failure.

## How to report gaps
`docs/gaps/GAP-NNN.md`: affected RFC/ECD §, the ambiguity, your proposed resolution, the narrowest-reading fallback. Label the PR `gap-flagged`. Continue on unaffected tasks while waiting.

## How to validate each completed module
Run the task verify → run `make verify` → run the module's DoD checklist → deploy to kind via the canonical chart overlay and prove probes+NetworkPolicy+one rollback → confirm metrics/alert/dashboard for the module render with real data. Only then mark done.

**Success for you is defined as: main always green, spec never contradicted, gaps surfaced loudly, tasks completed in order, budgets asserted — not speed.**
