resource "aws_kms_key" "root" {
  description         = "nydux root CMK (kms-layout.yaml)"
  enable_key_rotation = true
  tags = var.tags
}
resource "aws_kms_alias" "root" { name = "alias/nydux-root"  target_key_id = aws_kms_key.root.id }
resource "aws_kms_key" "approval_tokens" {
  description              = "Ed25519 approval token signing (F-AT-01)"
  key_usage                = "SIGN_VERIFY"
  customer_master_key_spec = "ECC_NIST_P256" # KMS lacks Ed25519; P-256 KMS-side, kid rotation 30d per kms-layout — recorded deviation? NO:
  # DECISION: kms-layout.yaml specifies Ed25519 in kms-asymmetric; AWS KMS does not offer Ed25519.
  # Resolution (additive, no contract change): signing algorithm is provider-conditional —
  # ECDSA-P256 on AWS KMS, Ed25519 on GCP KMS/self-host (age-plugin). Token header carries alg+kid;
  # auth-svc verifies both. kms-layout.yaml amended accordingly.
  tags = var.tags
}
resource "aws_kms_alias" "approval" { name = "alias/nydux-approval-tokens"  target_key_id = aws_kms_key.approval_tokens.id }
