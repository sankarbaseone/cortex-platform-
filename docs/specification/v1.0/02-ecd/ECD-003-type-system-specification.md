# ECD-003 — Type System Specification

**Level:** 2 (extends RFC-002, RFC-005 B.2/B.3, RFC-006 F.5/F.8, RFC-007, RFC-008). Types below are NORMATIVE names/fields; Claude Code transcribes them verbatim. Language mapping: proto = wire truth; Go domain structs mirror proto minus transport concerns; Rust/TS generated or mirrored per rules §3.12.

## 3.1 Global scalar & newtype rules
- `TenantID, ClusterID, RecID, PolicyID, BaselineID, JobID = uuid.UUID (v7)`; Go newtypes (`type TenantID uuid.UUID`) — prevents cross-assignment; constructors validate version=7.
- `KernelHash, FamilyHash = string` lowercase hex len 64, regex-validated at every boundary (proto validate + Go constructor + SQL CHECK — triple enforcement).
- Money: `type USDMicros int64` (ASM-002-2). Percent: `type Ratio float64` ∈[0,1]. Timestamps: `int64` epoch-nanos internal; RFC3339 at REST.
- All enums: proto enums with `_UNSPECIFIED=0`; Go mirrors; string forms SCREAMING_SNAKE.

## 3.2 Envelope & event payloads (proto — final wire types)
`EventEnvelope` frozen per RFC-005 B.2. Payload messages (package nydux.<domain>.v1), one per catalog row; representative full definitions (remaining rows follow identical field-style; every message enumerated in event-registry.yaml at Run 4 with these exact fields):
```proto
message KernelScored { string kernel_hash=1; string family_hash=2; string arch=3;
  double kes=4; map<string,double> components=5; string kes_model_version=6;
  double confidence=7; KernelStatus status=8; string toolchain_fp=9; }
enum KernelStatus { KERNEL_STATUS_UNSPECIFIED=0; SCORED=1; STATIC_ESTIMATE=2; MEASUREMENT_ONLY=3; }
message RegressionDetected { string family_hash=1; ToolchainRef from=2; ToolchainRef to=3;
  double delta_perf=4; double cri_contrib=5; repeated string mechanism_tags=6; string shape_class=7; }
message ToolchainRef { string fingerprint=1; string cuda=2; string ptxas=3; string triton=4; string torch=5; string host_cc=6; }
message CostCalculated { string slice_id=1; int64 usd_micros=2; string basis=3;
  int64 window_start_ns=4; int64 window_end_ns=5; string team_id=6; string model_id=7; string workload_id=8; }
message RecCreated { string rec_id=1; string kernel_hash=2; string pattern_id=3;
  double gain_p50=4; double gain_p90=5; double confidence=6; double risk=7;
  Evidence evidence=8; string patch_uri=9; }
message Evidence { repeated EvidenceRef refs=1; }  // ref: {kind: COUNTER|IR_SPAN|RUN|GRAPH, id, locator}
message GpuSample { string gpu_uuid=1; string node=2; string pod=3; string ns=4;
  float sm_util=5; float mem_util=6; uint32 mem_used_mb=7; float power_w=8; float temp_c=9;
  uint32 sm_clock=10; float pcie_tx_mb=11; float nvlink_tx_mb=12; float tensor_active=13;
  float dram_active=14; uint32 xid=15; }
```
Validation: buf protovalidate rules on every message (hash regex, ranges kes∈[0,100], confidence∈[0,1], usd_micros≥0 except savings deltas).

## 3.3 Domain aggregates (Go; internal/domain)
```go
// kernel-registry
type Kernel struct { Hash KernelHash; Family FamilyHash; Tenant TenantID; Arch Arch;
  Status KernelStatus; KES *Score; FirstSeen, LastSeen time.Time }
type Score struct { Value float64; Components map[Component]float64;
  ModelVersion string; Confidence float64 }
type Component string // enum: roof,occ,stall,mem,tc,mix — closed set, validated
// legal status transitions (register.go enforces):
// MEAS_ONLY→STATIC→SCORED (forward only); SCORED→SCORED (rescore); any→MEAS_ONLY on parser-loss forbidden (history kept, new row semantics)
type Toolchain struct { ID uuid.UUID; Tenant TenantID; Ref ToolchainRef;
  Approval ApprovalState; ApprovedBy *UserID; ApprovedAt *time.Time }
type ApprovalState string // unreviewed|approved|revoked; transitions: unreviewed→approved|revoked; approved→revoked; revoked terminal
```
```go
// regression-svc
type RunStats struct { N int; MedianNs int64; MAD float64 } // Welford+P2 quantile internally
type Comparison struct { Family FamilyHash; From, To ToolchainRef; Arch Arch;
  ShapeClass string; Delta float64; Sigma float64; Regressed bool }
type FleetCRI struct { Window TimeRange; Value float64; Contribs []Comparison } // Σ contrib, RFC-002 2.6
type GateExpr struct { Metric GateMetric /*CRI only V1*/; Op CmpOp; Threshold float64 }
```
```go
// recommender
type Recommendation struct { ID RecID; Tenant TenantID; Kernel KernelHash; Pattern PatternID;
  Gain GainDist; Confidence float64; Effort EffortClass; Risk float64;
  State RecState; Evidence Evidence; PatchURI *string; Decided *Decision }
type GainDist struct{ P50, P90 float64 }
type RecState string // created→approved→applied→verified | created→rejected | applied→failed→rolled_back
// state machine table is the single source; transitions elsewhere are compile-error (unexported setters)
type PatternID string // closed set patterns 1..12 per RFC-002 2.7: "unfused_elementwise","missing_flashattn",
// "reg_pressure","noncoalesced","fp32_gemm_tc","smallbatch_gemm","launch_storm","triton_tile","cublas_fallback",
// "comm_overlap_gap","toolchain_upgrade","toolchain_holdback"
```
```go
// savings-svc
type Baseline struct { ID BaselineID; Tenant TenantID; Version int; FrozenState FrozenStateRef; // blob hash
  TwinModelVersion string; AnchoredAt time.Time; Cosign Cosign }
type Cosign struct{ Customer, Nydux *SignatureRef }
type SavingsPeriod struct { Period YearMonth; Total USDMicros; Method MethodDual; Actions []ActionShare }
type MethodDual struct{ Twin, Trailing USDMicros; Contractual USDMicros /* = min */ }
type ActionShare struct{ RecID RecID; Shapley USDMicros; CI *Interval /*nil when exact*/ }
```
```go
// twin-svc
type Scenario struct { Base SnapshotRef; Mut []Mutation } // Mutation = oneof {GpuSku, Parallelism, ToolchainRef, Quantization, NodeCount}
type Parallelism struct{ TP, PP, DP, EP, MicroBatch int } // all ≥1; TP*PP*DP consistency validated vs NodeCount*GPUsPerNode
type Prediction struct { Throughput, TTFT, TPOT, PowerW float64; CostUSDPerHour USDMicros;
  Band Interval; Support SupportLevel } // SupportLevel: FULL|LOW_SUPPORT
```
```go
// policy/audit
type PolicyDoc struct{ ID PolicyID; Rego string; Enforcement Enforcement /*block|warn|audit*/; Version int }
type Decision struct{ Policy PolicyID; Verdict Verdict; Subject json.RawMessage; LedgerSeq int64 }
type AuditEntry struct{ Seq int64; Prev, Hash [32]byte; Actor string; Action string; Subject CanonicalJSON; At time.Time }
type CanonicalJSON []byte // constructor sorts keys, rejects floats (ASM-002-2), NFC strings
```

## 3.4 Rust core types (analyzers)
```rust
pub struct CanonicalIr { pub dialect: Dialect, pub text: String, pub meta: Meta } // Meta{arch, versions}
pub enum Dialect { TtIr, TtgIr, LlvmIr, Ptx, FxText, Hlo }
pub struct PtxFeatures { pub regs: u16, pub instr_mix: InstrMix, pub mem_ops: MemOps, pub barriers: u16 }
pub struct SassFeatures { pub spills: u16, pub occ_inputs: OccInputs, pub decode_confidence: f32 }
pub struct KesInput { pub static_: StaticFeatures, pub dynamic: Option<NcuCounters>, pub profile: Profile }
pub enum Profile { Training, Prefill, Decode, Hpc }
// COMMUTATIVE_OPS (canonicalizer, closed set V1): add,mul,and,or,xor,max,min,fadd,fmul (float variants only when reassoc flag present)
// family mask groups (hash.rs): {tile_m,tile_n,tile_k,num_warps,num_stages,BLOCK_*} — masked for family_hash
```
Thread safety: parser structs `!Sync` by design (ECD-002 §2); shared caches behind `Arc<RwLock<...>>` only in cache crate. Memory: bounded arenas per analysis (cap 512MiB, config), overflow ⇒ typed error → MEASUREMENT_ONLY.

## 3.5 KES weight tables (normative defaults; calibrate.rs may refit, versions bump kes_model_version)
| Profile | roof | occ | stall | mem | tc | mix |
|---|---|---|---|---|---|---|
| training | .30 | .15 | .20 | .15 | .10 | .10 |
| prefill  | .30 | .15 | .20 | .15 | .10 | .10 |
| decode   | .25 | .15 | .20 | .30 | .00* | .10 |  *tc weight redistributed per RFC-002 2.5 edge-case
| hpc      | .35 | .15 | .20 | .15 | .05 | .10 |
Static-only confidence: base .5, −.1 if decode_confidence<.6, +.1 if NCU-free but DCGM-coarse present; clamp [.4,.6].

## 3.6 Verify tolerances (bench-runner numerical-equivalence; per dtype)
fp32: rtol 1e-5 / atol 1e-6 · bf16: 1.6e-2 / 1e-3 · fp16: 1e-3 / 1e-4 · fp8_e4m3: 6e-2 / 1e-2 · int8: exact. Reductions compared after deterministic-order reference run. Failure ⇒ rec cannot leave `created`.

## 3.7 API DTOs (REST; TS + pydantic generated from OpenAPI — never hand-written)
Response wrappers: list = `{items: T[], next_cursor?: string}`; errors = problem+json `{type,title,status,detail,trace_id,errors?[]}` where `type` ∈ error catalog URIs `https://errors.nydux.ai/<code>` (catalog: `validation`, `not_found`, `conflict`, `rate_limited`, `gate_failed`, `policy_blocked`, `unsupported_ir`, `low_support`, `unauthorized`, `forbidden`, `internal` — closed set V1; each mapped to HTTP+gRPC code in libs/go/nyxproblem table).

## 3.8 Repository & service interfaces (ports; one exemplar per store kind — every service's ports.go uses these shapes with its aggregates)
```go
type KernelRepo interface {
  Upsert(ctx context.Context, k Kernel) error                 // natural-key idempotent
  Get(ctx context.Context, t TenantID, h KernelHash) (Kernel, error) // ErrNotFound sentinel
  List(ctx context.Context, t TenantID, q ListQuery) (Page[Kernel], error) // keyset cursor
}
type Bus interface { Publish(ctx context.Context, e Envelope) error } // outbox-backed impl in adapters
type Clock interface{ Now() time.Time }
type Flags interface{ Bool(ctx context.Context, key string, def bool) bool }
```
`ListQuery{Filter rsql.Expr; Sort []SortKey; Limit int; Cursor Cursor}` — rsql grammar closed to fields whitelisted per endpoint (whitelists enumerated in api-registry.yaml Run 4; interim: exactly the fields shown in RFC-006 examples).

## 3.9 Agent framework types (Python, pydantic v2; RFC-008)
```python
class Task(BaseModel): id: UUID7; tenant: UUID7; kind: AgentKind; input: TaskInput
  budget: Budget; state: TaskState  # created|planning|executing|awaiting_approval|done|failed
class Budget(BaseModel): max_input_tokens: int; max_output_tokens: int; max_tool_calls: int; deadline_s: int
class ToolCall(BaseModel): tool: str; args: dict; args_hash: str; result_ref: str | None
class JudgeVerdict(BaseModel): passed: bool; grounding: float  # must ==1.0 for pass (H.9)
  reasons: list[str]
class AgentSpec(BaseModel):  # one const instance per domains/<agent> (ECD-002 §5)
  kind: AgentKind; tools: list[ToolScope]; triggers: list[Trigger]; judge_extras: list[str]; writes: WriteClass
class WriteClass(str, Enum): none="none"; rec_draft="rec_draft"; policy_exec="policy_exec"; calib="calib"
```
Serialization: model_dump(mode="json") only; no pickle anywhere (security rule).

## 3.10 Configuration objects
Every service: `Config{HTTPOps, GRPC, PG?, CH?, Kafka?, Redis?, OTel, Flags, SvcSpecific}` sub-structs from libs/go; each field: env name, default, validation range documented inline. Analyzer configs mirrored in Rust `serde` structs with `deny_unknown_fields`.

## 3.11 Lifecycle & ownership rules
Aggregates constructed only via `New*` funcs (invariants enforced at construction); repos return copies (no shared mutable state across goroutines); events built from aggregates via `ToEvent()` methods living beside the aggregate (single mapping point). Rust: analysis outputs are owned values moved into gRPC responses; no `'static` borrows of inputs.

## 3.12 Cross-language mirroring rule
Proto is the wire truth. Go domain types may differ from proto ONLY by: newtypes, time.Time, decoded enums. TS/Python client types are generated. Any third representation (e.g., DB row structs via sqlc) maps through explicit `toDomain/fromDomain` functions with round-trip property tests.

## Assumptions
ASM-003-1: KES weight numeric defaults (§3.5) instantiate RFC-002 "default weights" + decode-profile redistribution; calibrate refits per §2.5 procedure. ASM-003-2: verify tolerances (§3.6) instantiate RFC-008 "numerical equivalence" — values chosen from standard mixed-precision testing practice; founders may tighten per customer contract. ASM-003-3: error catalog (§3.7) closed-set instantiates RFC-006 "error catalog" reference.
