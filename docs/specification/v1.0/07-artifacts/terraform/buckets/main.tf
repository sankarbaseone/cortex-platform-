locals { buckets = { walg = "walg", audit_anchor = "audit-anchor", replay = "replay-bundles", ch_cold = "ch-cold" } }
resource "aws_s3_bucket" "b" { for_each = local.buckets  bucket = "${var.prefix}-${each.value}"  tags = var.tags }
resource "aws_s3_bucket_server_side_encryption_configuration" "b" {
  for_each = aws_s3_bucket.b
  bucket = each.value.id
  rule { apply_server_side_encryption_by_default { sse_algorithm = "aws:kms", kms_master_key_id = var.kms_key_arn } }
}
resource "aws_s3_bucket_versioning" "b" {
  for_each = aws_s3_bucket.b
  bucket = each.value.id
  versioning_configuration { status = "Enabled" }
}
resource "aws_s3_bucket_object_lock_configuration" "anchor" {
  bucket = aws_s3_bucket.b["audit_anchor"].id
  rule { default_retention { mode = "COMPLIANCE", days = 2555 } }  # 7y (RFC-007 retention)
}
resource "aws_s3_bucket_lifecycle_configuration" "b" {
  for_each = aws_s3_bucket.b
  bucket = each.value.id
  rule { id = "tiering", status = "Enabled"
    transition { days = 30, storage_class = "STANDARD_IA" } }
}
resource "aws_s3_bucket_public_access_block" "b" {
  for_each = aws_s3_bucket.b
  bucket = each.value.id
  block_public_acls = true  block_public_policy = true
  ignore_public_acls = true restrict_public_buckets = true
}
