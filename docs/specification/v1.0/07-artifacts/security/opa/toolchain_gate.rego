# policy-svc built-in: toolchain deploy gate — only approved toolchains reach prod (ECD-012 §3).
package nydux.toolchain_gate
import rego.v1
default allow := false
allow if {
  input.subject_kind == "toolchain"
  input.toolchain.approval == "approved"
  not fleet_gate_failed
}
fleet_gate_failed if input.fleet_cri > data.limits.max_cri   # default 0.10
deny_reason contains "toolchain_unreviewed" if input.toolchain.approval != "approved"
deny_reason contains "cri_gate_failed" if fleet_gate_failed
