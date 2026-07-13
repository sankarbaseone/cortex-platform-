# policy-svc built-in: recommendation apply guard (ECD-012 §3). Package name is contract.
package nydux.rec_guard

import rego.v1

default allow := false

# input: {subject_kind:"recommendation", rec:{state, risk, confidence, pattern_id},
#         actor:{user_id, roles[]}, token:{user_id, rec_id, exp_ns}, now_ns}

allow if {
  input.subject_kind == "recommendation"
  input.rec.state == "approved"
  token_valid
  not author_is_approver
  input.rec.risk <= data.limits.max_auto_risk        # tenant-tunable, default 0.5
}

token_valid if {
  input.token.rec_id == input.rec.rec_id
  input.token.exp_ns > input.now_ns
}

author_is_approver if input.rec.author_id == input.token.user_id

deny_reason contains "rec_not_approved" if input.rec.state != "approved"
deny_reason contains "token_expired" if input.token.exp_ns <= input.now_ns
deny_reason contains "sod_violation" if author_is_approver
deny_reason contains "risk_exceeds_limit" if input.rec.risk > data.limits.max_auto_risk
