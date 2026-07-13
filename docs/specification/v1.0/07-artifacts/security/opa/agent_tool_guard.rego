# agent-orchestrator sidecar policy: write-class tools require human approval evidence (RFC-008 H.5).
package nydux.agent_tool_guard
import rego.v1
default allow := false
allow if { not input.tool.write_class }                       # read tools: always
allow if { input.tool.write_class; input.approval_token_valid; input.policy_verdict == "allow" }
deny_reason contains "write_without_token" if { input.tool.write_class; not input.approval_token_valid }
deny_reason contains "policy_not_allow" if { input.tool.write_class; input.policy_verdict != "allow" }
