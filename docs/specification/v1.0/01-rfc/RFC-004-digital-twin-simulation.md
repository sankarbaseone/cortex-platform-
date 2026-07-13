# RFC-004 — Digital Twin & Simulation Engine

**Status:** Approved · **Extends:** V1 Phase 4.9, 7.4 · **Owns Section E:** Digital Twin, Simulation, Savings Attribution, Forecasting, Capacity Planning, Scheduling model, Anomaly Detection, Failure Prediction, Risk Scoring.

## 4.1 Twin Model
State = (Cluster topology graph, GpuSku params, Model arch spec, parallelism P={tp,pp,dp,ep,microbatch}, Toolchain, workload trace class). Output = {throughput, latency{ttft,tpot}, cost, power, utilization} with uncertainty bands.

**Two-level model:**
1. **Analytical core (roofline+):** per-op time `t_op = max(F/peak_f·η_f, B/peak_b·η_b)`; collective time from α-β model `t_coll = α·steps(algo,topo) + β·bytes/bw_bisection`; pipeline schedule simulated (1F1B) for bubbles; overlap modeled as `max(compute, comm)` per stage when overlap flag set else sum.
2. **Learned residual:** GBM (LightGBM) predicting `log(t_measured/t_analytical)` from features {op mix, KES of hot kernels, arch, P, seq-len class, toolchain}; trained on Knowledge Graph MeasurementRuns. Rationale (V1): roofline alone ignores kernel quality/CPU overhead; residual closes the gap and improves with the moat.
**Accuracy SLO:** MAPE ≤15% on holdout by GA; per-prediction band = residual model quantiles (P10/P90). Predictions below data-support threshold (nearest-neighbor distance in feature space > τ) are labeled LOW_SUPPORT and widened.
**Complexity:** analytical O(ops·stages); one scenario <100ms; sweeps parallelized.

## 4.2 Simulation Engine
API: submit ScenarioSet {base_state, mutations[]} → grid/latin-hypercube sweep → Pareto set (cost vs throughput vs latency). Deterministic seed; results cached by scenario hash. Used by: capacity planning, what-if UI, savings baselines, Capacity/Twin agents.

## 4.3 Savings Attribution (patent #7, OQ-03/OQ-13)
**Baseline B:** twin prediction of {cost/token or cost/step} for the frozen pre-engagement state (topology, toolchain, kernels), replayed against ACTUAL demand each period ⇒ removes demand-mix gaming.
**Savings S(period) = Σ demand_t · (unit_cost_B,t − unit_cost_actual,t)**, floored at 0 per line-item; dual-reported next to naive trailing-30d baseline; contract takes the lower (OQ-03 interim).
**Attribution to actions:** Shapley over the action set A applied in period (each action's marginal effect measured by twin with action toggled); |A|≤10 enforced by grouping, exact Shapley O(2^|A|) acceptable; else Monte-Carlo permutations (m=200, CI reported).
**Re-anchoring (OQ-13):** model/infra change events ≥ threshold trigger baseline re-anchor requiring customer co-signature (UI flow RFC-010; audit-logged).
**Edge cases:** negative savings shown (honesty builds trust, per V1 principles); missing telemetry windows excluded from both sides symmetrically.

## 4.4 Forecasting
Demand (tokens, GPU-hours) per workload: seasonal model (Prophet-class: trend+weekly+holiday) + burst overlay from queue-depth signals; horizon 1–13 weeks; P50/P90. Cost forecast = demand forecast × unit-cost forecast (unit cost from twin under planned changes). Backtest monthly; publish MAPE per tenant.

## 4.5 Capacity Planning
ILP: choose node counts n_s per SKU s and parallelism P to minimize `Σ n_s·rate_s` s.t. twin-predicted throughput ≥ P90 demand, latency SLO satisfied, power/rack constraints, commit contracts honored. Solver: OR-Tools CBC; problem sizes tiny (<10^3 vars). Output includes sensitivity (shadow prices) — "one more H100 buys X tok/s".

## 4.6 Scheduling model (advisory)
Bin-packing recommendation for fractional/MIG allocation: first-fit-decreasing on (mem, SM%) vectors with anti-affinity constraints; interacts with Run:ai-class schedulers as ADVICE ONLY (V1 Phase 13: complementary, not competing).

## 4.7 Multi-tenant attribution model (OQ-06)
Per-request GPU-seconds ≈ Σ_phase tokens_r,phase · w_phase where w learned per (model, arch, batch regime) from isolated calibration runs; published error bound (target ±10%); disclosed in methodology doc customers sign.

## 4.8 Anomaly Detection
Per-metric seasonal-robust z-score (median/MAD over matched weekly seasonality) + CUSUM for drifts; cost anomalies additionally gated by min-$ threshold to avoid noise. Alerts carry top correlated changes (toolchain deploys, config changes) from the change-event stream ⇒ feeds RCA agent.

## 4.9 Failure Prediction (patent #20)
Features: XID history, ECC correctable rate slope, thermal excursions, fan/power variance. Model: gradient-boosted survival (predict hazard within 7d). Action: alert + recommended drain (human-approved; Policy Engine gate). Precision-first threshold (false drain is expensive): target precision ≥0.8 at whatever recall results; report both.

## 4.10 Risk Scoring (for recommendations & policies)
`risk = severity(action_class) · blast_radius(scope) · (1−confidence) · env_weight(prod=3,staging=1)`; classes: read=0, config=1, workload-restart=2, toolchain-change=3, quantization=3. Score >θ ⇒ mandatory dual approval (RFC-009 §I.6).

## 4.11 Testing
Twin: golden scenarios with published measured results (internal cluster + public MLPerf-style configs) — CI asserts MAPE ceiling. Savings: property tests (symmetry, demand-invariance under baseline replay, Shapley efficiency Σφ = total). Forecast: rolling-origin backtests in CI weekly job.

## 4.12 Failure Modes
Residual model unavailable ⇒ analytical-only with widened bands + LOW_SUPPORT label. Baseline dispute ⇒ full replay bundle exportable (inputs, model versions, seeds) so customer can reproduce — determinism is a contractual feature.
