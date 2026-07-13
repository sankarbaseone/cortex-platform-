# RFC-006 — API & SDK Specification

**Status:** Approved · **Extends:** V1 Phase 5.2 · **Owns Section F** entirely.

## F.1 Surfaces
- **Public REST** `https://api.nydux.ai/v1` — CRUD, queries, approvals. OpenAPI 3.1 is the contract (excerpt F.8); full spec generated from protobuf annotations (grpc-gateway) so REST and gRPC never drift.
- **gRPC** `grpc.nydux.ai:443` — streaming ingest (edge-gateway→CP), inter-service, high-volume reads.
- **SSE** on REST endpoints with `Accept: text/event-stream` for live dashboards; **WebSocket** only `/v1/graph/ws` for Graph Explorer (OQ-15).
- **GraphQL** internal BFF only (OQ-10), read-only, not a supported public contract.
- **CLI** `nydux` (Go, cobra) and **SDK** (Python primary; Go/TS generated).

## F.2 AuthN/AuthZ
OIDC (Auth0/Keycloak-compatible); humans: Authorization: Bearer JWT (15-min access, refresh rotation). Machines: OAuth2 client-credentials per service account; DP gateways: mTLS client certs + bound JWT (SPIFFE ID in SAN). Every token carries tenant_id + roles; API layer enforces coarse RBAC, services enforce fine ABAC (RFC-009 §I.5). API keys (hashed, prefix-identifiable `nydx_live_…`) allowed for CI plugin only, scope-limited to `ci:regressions`.

## F.3 Cross-cutting REST semantics
- Pagination: cursor-based `?limit=100&cursor=…` (opaque, base64 of keyset); no offset pagination anywhere.
- Filtering: RSQL-subset `?filter=kes<50;arch==sm_90` documented per endpoint; sorting `?sort=-kes`.
- Errors: RFC 9457 `application/problem+json` with `type` URI in our error catalog, `trace_id` always included.
- Idempotency: all POSTs accept `Idempotency-Key` (stored 24h).
- Rate limits: token bucket per (tenant, principal): default 600 r/min read, 120 r/min write; headers `RateLimit-*`; 429 with `Retry-After`.
- Timeouts: server 30s REST (long ops become async jobs `202 + /v1/jobs/{id}`); gRPC deadlines mandatory, default 10s, streaming keepalive 30s.
- Retry guidance: clients retry idempotent verbs on 429/502/503/504 with exp backoff + jitter, max 4; SDK implements this.
- Versioning: URI major (`/v1`); additive changes only within major; `Deprecation` + `Sunset` headers 12 months before removal.

## F.4 Endpoint map (public v1)
```
/kernels                      GET list · /kernels/{hash} GET · /kernels/{hash}/score GET
/kernels/{hash}/ir/{stage}    GET (in-tenant deployments only; SaaS returns 404 by design)
/kernels/{hash}/recommendations GET
/regressions                  GET · POST /regressions/checks (CI gate) 
/toolchains                   GET/POST · /toolchains/{id}/approval PUT (governance)
/recommendations              GET · /{id} GET · /{id}/approve POST · /{id}/reject POST · /{id}/apply POST
/simulations                  POST (ScenarioSet) → 202 job · /simulations/{id} GET
/capacity/plans               POST/GET
/finance/attribution          GET (dims: team|model|kernel|token) 
/finance/savings              GET · /finance/baselines POST/GET · /{id}/reanchor POST (cosign)
/clusters /devices /rates     GET (infra)
/policies                     CRUD · /policies/{id}/decisions GET
/audit/entries                GET (auditor role) · /audit/verify POST (chain check)
/agents/tasks                 GET · /{id} GET (transcript refs)
/tenants /users /roles /apikeys  admin CRUD
/jobs/{id}                    GET async job status
```

## F.5 gRPC (protobuf excerpt — canonical style)
```proto
syntax = "proto3"; package nydux.compiler.v1;
service KernelService {
  rpc GetKernel(GetKernelRequest) returns (Kernel);
  rpc ListKernels(ListKernelsRequest) returns (ListKernelsResponse);
  rpc StreamScores(StreamScoresRequest) returns (stream KernelScore); // server-stream
  rpc SubmitArtifacts(stream ArtifactChunk) returns (SubmitAck);      // client-stream (gateway ingest)
}
message Kernel { string kernel_hash=1; string family_hash=2; string arch=3;
  KernelScore score=4; repeated string pattern_tags=5; Status status=6; }
message KernelScore { double kes=1; map<string,double> components=2;
  string kes_model_version=3; double confidence=4; int64 scored_at_ns=5; }
```
Interceptors (all services): auth, tenant-context, deadline-propagation, otel-trace, panic-recovery, rate-limit. Errors: `google.rpc.Status` + `ErrorInfo{reason, domain="nydux.ai", metadata{trace_id}}`.

## F.6 SDK (Python, primary)
```python
import nydux
cli = nydux.Client()                      # env NYDUX_API_KEY / OIDC
@nydux.profile(model="stt-prod")          # V1 decorator pattern
def train_step(...): ...
cli.kernels.list(filter="kes<50", sort="-cost")
gate = cli.regressions.check(from_tc="cuda12.4", to_tc="cuda12.6", fail_on_cri=0.10)
sim = cli.simulations.run(scenarios).wait()
```
Design rules: sync + async clients; typed models (pydantic) generated from OpenAPI; built-in retry/idempotency; zero required config in-cluster (workload identity autodetect); no raw IR ever uploaded by SDK in SaaS mode (privacy guard hard-coded, RFC-003 §D.9).

## F.7 CLI
`nydux login|scan|kernel score <hash>|kernels top|regressions [--fail-on CRI>x] [--junit out.xml]|simulate -f scenario.yaml|savings report --period 2026-06|policy test -f policy.rego|dlq …(admin)`. Exit codes: 0 ok, 1 error, 2 gate-failed (CI-friendly). Output: table default, `-o json|yaml`.

## F.8 OpenAPI excerpt (normative fragment)
```yaml
openapi: 3.1.0
info: {title: NYDUX API, version: 1.0.0}
paths:
  /v1/kernels/{hash}/score:
    get:
      operationId: getKernelScore
      parameters: [{name: hash, in: path, required: true, schema: {type: string, pattern: "^[a-f0-9]{64}$"}}]
      responses:
        "200": {content: {application/json: {schema: {$ref: "#/components/schemas/KernelScore"}}}}
        "404": {$ref: "#/components/responses/Problem"}
components:
  schemas:
    KernelScore:
      type: object
      required: [kes, components, kesModelVersion, confidence]
      properties:
        kes: {type: number, minimum: 0, maximum: 100}
        components: {type: object, additionalProperties: {type: number}}
        kesModelVersion: {type: string}
        confidence: {type: number, minimum: 0, maximum: 1}
```
Example payload:
```json
{"kes": 41.7, "components": {"roof":0.38,"occ":0.71,"stall":0.55,"mem":0.34,"tc":0.12,"mix":0.9},
 "kesModelVersion":"kes-2026.07-sm90","confidence":0.93}
```

## F.9 Testing
Contract tests: schemathesis fuzz on OpenAPI in CI; grpc buf-breaking check; SDK integration suite against ephemeral stack (RFC-011); rate-limit and pagination property tests.
