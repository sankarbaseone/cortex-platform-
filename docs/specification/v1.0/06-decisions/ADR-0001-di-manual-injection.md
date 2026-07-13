# ADR-0001: Manual constructor injection, wired in main.go only
Status: accepted · Date: 2026-07-13 · Owner: platform
## Context
16 Go services need a uniform DI approach (ECD-004 §4.1 decision, recorded as ADR for discoverability).
## Decision
Manual constructor injection; all wiring in cmd/<svc>/main.go via nyxrun.Group; fixed order config→otel→stores→bus→repos→domain→transports.
## Alternatives rejected
google/wire (codegen step, no benefit at this scale); uber/fx (runtime lifecycle magic, conflicts with crash-only design); service locator (hides dependency graph from depguard).
## Consequences
Compile-time failures instead of container errors; slightly longer main.go; exemplar in services/kernel-registry.
## Compliance
Implements ECD-004 §4.1; touches no frozen contract.
