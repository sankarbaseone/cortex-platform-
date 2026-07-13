package nydux.rec_guard_test
import rego.v1
import data.nydux.rec_guard

base := {"subject_kind": "recommendation", "now_ns": 100,
  "rec": {"rec_id": "r1", "state": "approved", "risk": 0.2, "author_id": "u-author"},
  "token": {"rec_id": "r1", "user_id": "u-approver", "exp_ns": 200}}

test_allow_happy if rec_guard.allow with input as base with data.limits.max_auto_risk as 0.5
test_deny_expired if not rec_guard.allow with input as object.union(base, {"now_ns": 300}) with data.limits.max_auto_risk as 0.5
test_deny_sod if not rec_guard.allow with input as object.union(base, {"token": {"rec_id": "r1", "user_id": "u-author", "exp_ns": 200}}) with data.limits.max_auto_risk as 0.5
test_deny_risk if not rec_guard.allow with input as object.union(base, {"rec": object.union(base.rec, {"risk": 0.9})}) with data.limits.max_auto_risk as 0.5
test_deny_unapproved if not rec_guard.allow with input as object.union(base, {"rec": object.union(base.rec, {"state": "created"})}) with data.limits.max_auto_risk as 0.5
