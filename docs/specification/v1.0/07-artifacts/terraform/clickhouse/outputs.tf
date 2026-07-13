output "clickhouse_host"   { value = "clickhouse-nydux.${var.namespace}.svc" }
output "backup_bucket_arn" { value = aws_s3_bucket.backup.arn }
output "node_group_name"   { value = aws_eks_node_group.analytics.node_group_name }
