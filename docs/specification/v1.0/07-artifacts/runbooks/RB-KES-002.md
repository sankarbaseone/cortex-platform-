# RB-KES-002 — NyduxAnalyzerMeasurementOnlySpike (ticket)
**Trigger:** >20% kernels MEASUREMENT_ONLY — parser matrix gap (OQ-05 risk). · **Owner:** on-call

## Symptoms
>20% kernels MEASUREMENT_ONLY — parser matrix gap (OQ-05 risk).

## Dashboards
NYDUX / Service Overview (+ domain dashboard per dashboard-registry).

## Diagnosis tree
1. Which source/toolchain? group kernel_events by toolchain_fp for MEAS_ONLY.
2. New Triton/CUDA release in the wild? Confirm against toolchains table approvals.
3. Decode_confidence low on SASS only: sass-decoder arch gap (new sm_).

## Mitigation
Ship parser-matrix entry per RFC-002 process (goldens first); until then customers see measurement-only scores — set expectations in release notes.

## Post-actions
Matrix coverage table updated; add the toolchain to canary corpus.
