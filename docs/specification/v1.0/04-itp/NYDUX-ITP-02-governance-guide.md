# NYDUX Development Governance Guide — the Engineering Constitution

## 1. Authority hierarchy (absolute, unchanged)
Founder (via ADR) → RFCs → ECDs → machine registries → implementation code. Lower layers may never contradict higher ones; conflicts halt work and go up, never sideways.

## 2. ADR workflow (the ONLY way the spec changes)
Trigger: reality contradicts a frozen document (library API changed, provider limitation, measured budget impossible) or a GAP escalates. Process: (1) author ADR from `docs/adr/ADR-0000-template.md` — context, decision, rejected alternatives, consequences, compliance section listing every touched RFC/ECD clause; (2) label PR `adr-proposal`; (3) founder approves/rejects — no lazy consensus, explicit approval only; (4) on approval, the ADR is appended (never edited into) the corpus, registries updated in the same PR, spec version bumps 1.0 → 1.0.x (additive) or 1.x (contract-touching); (5) `spec-manifest.sha256` regenerated. Precedents: ADR-0001 (DI), ADR-0002 (outbox), and the PRR F-1 KMS-curve amendment.

## 3. GAP protocol (ambiguity channel)
When implementation hits genuine ambiguity: STOP the affected task → create `docs/gaps/GAP-NNN.md` (template: affected RFC/ECD §, the ambiguity, proposed resolution, narrowest-reading fallback) → mark PR `gap-flagged` → founder decides within the sprint: (a) clarification comment (no spec change), (b) spawn ADR, or (c) reject task interpretation. Work MAY continue on unaffected tasks; the gapped task waits for founder approval. A GAP is never resolved silently in code.

## 4. Specification versioning
v1.0 = freeze. Patch (1.0.x): additive clarifications via ADR. Minor (1.x): contract-touching ADR (proto field additions, new endpoints). Major (2.0): new RFC cycle — out of scope for this program. Every release notes the spec version it implements.

## 5. Change approval matrix
| Change class | Approval |
|---|---|
| Implementation code within a build-order task | CI gates + self-merge on green (solo phase); +1 reviewer once team ≥2 |
| New dependency | `deps-approval` label → founder |
| Registry file change | founder (contract) — CI blocks unlabeled registry diffs |
| proto / migration / OpenAPI / chart shape | ADR only |
| GAP resolution | founder |
| Emergency hotfix | expedited train — gates NOT skipped (RFC-011 J.6), post-hoc review ≤48h |

## 6. Engineering review process
Every PR: CI quality gates Q (non-negotiable) + conventional commit with task ID + DoD checklist from ITP-05 pasted in description. Contract-adjacent paths (proto/, registries/, migrations) require the founder as reviewer even solo (self-review with 24h cooling period). Weekly: review `gap-flagged` and `adr-proposal` queues; monthly: drill + flake + error-budget review (RFC-011 J.10/J.11).

## 7. Merge rules
Trunk-based, main always releasable, squash-merge, linear history, no direct pushes to main (branch protection from T-001), revert-first policy on red main (fix forward only with founder call).
