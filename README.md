# NYDUX Platform
Compiler-aware Enterprise AI Operating System — control plane, data plane, and web app monorepo.

## What is this
NYDUX observes GPU fleets at the compiler level (Triton/Inductor/PTX/SASS), scores every kernel (KES), detects toolchain regressions (CRI), attributes cost to kernels and teams, simulates changes on a digital twin, and drives human-approved optimizations — with a cryptographic audit chain end to end. See `docs/rfcs/` for the frozen architecture and `docs/ecds/` for construction documents.

## Repo map (ECD-001)
`services/` 16 Go services · `crates/` Rust analyzers (canonicalizer, ptx-parser, sass-decoder, kes) · `dataplane/` collector + edge-gateway + analyzers · `proto/` wire truth · `web/app` React UI · `libs/` shared go/ts/py · `charts/`, `infra/` deploy · `registries/` machine-readable source of truth · `tools/` codegen + CI checks.

## Quickstart (dev)
Prereqs: mise (pins Go 1.23.4/Rust 1.80.1/Python 3.12.6/Node per .tool-versions), Docker, kind, buf.
```
make bootstrap   # toolchains, hooks, buf deps
make dev-up      # compose deps + kind + migrations + seed
make gen         # protos, sqlc, openapi clients
make test        # unit + property, race on
make verify      # full pre-PR gate
make e2e SUITE=smoke
```
Service dev loop: `make run SVC=kernel-registry` (compose deps must be up).

## Contributing
Trunk-based; PRs to main; quality gates in `.github/workflows/ci.yaml` are non-negotiable (lint, tests, buf breaking, registry drift, privacy egress, security scans). Commit style and rules: `CLAUDE.md` (humans too). Codeowners: `registries/ownership.yaml` renders CODEOWNERS.

## Security
Report vulnerabilities to security@nydux.ai. Architecture: RFC-009; threat model: `docs/security/threat-model.md`. Never commit secrets — CI blocks, ESO delivers at runtime.

## License
Proprietary — © NYDUX Pvt Ltd. All rights reserved.
