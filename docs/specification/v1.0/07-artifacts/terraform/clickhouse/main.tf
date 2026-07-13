# ClickHouse on dedicated NVMe node group + Altinity operator, per RFC-001 A.3/A.10, ECD-006.
resource "aws_eks_node_group" "analytics" {
  cluster_name    = var.cluster_name
  node_group_name = "nydux-analytics"
  instance_types  = [var.instance_type]
  scaling_config { desired_size = var.shards * var.replicas, min_size = var.shards * var.replicas, max_size = var.shards * var.replicas + 2 }
  labels = { "nydux.ai/pool" = "analytics" }
  taint { key = "nydux.ai/analytics", value = "true", effect = "NO_SCHEDULE" }
  tags = var.tags
}

resource "helm_release" "ch_operator" {
  name       = "clickhouse-operator"
  repository = "https://docs.altinity.com/clickhouse-operator"
  chart      = "altinity-clickhouse-operator"
  version    = "0.24.0"
  namespace  = var.namespace
  create_namespace = true
}

resource "kubernetes_manifest" "chi" {
  manifest = {
    apiVersion = "clickhouse.altinity.com/v1"
    kind       = "ClickHouseInstallation"
    metadata   = { name = "nydux", namespace = var.namespace }
    spec = {
      configuration = {
        clusters = [{
          name   = "nydux"
          layout = { shardsCount = var.shards, replicasCount = var.replicas }
        }]
        zookeeper = { nodes = [for i in range(var.keeper_replicas) : { host = "keeper-${i}.keeper.${var.namespace}" }] }
        settings  = { "storage_configuration/disks/cold/type" = "s3",
                      "storage_configuration/disks/cold/endpoint" = "https://${var.cold_bucket}.s3.amazonaws.com/cold/" }
      }
      templates = {
        podTemplates = [{
          name = "ch-pod"
          spec = {
            nodeSelector = { "nydux.ai/pool" = "analytics" }
            tolerations  = [{ key = "nydux.ai/analytics", operator = "Equal", value = "true", effect = "NoSchedule" }]
            containers   = [{ name = "clickhouse", image = "clickhouse/clickhouse-server:24.8",
                              resources = { requests = { cpu = "4", memory = "24Gi" }, limits = { cpu = "8", memory = "24Gi" } } }]
          }
        }]
        volumeClaimTemplates = [{
          name = "data"
          spec = { accessModes = ["ReadWriteOnce"], resources = { requests = { storage = "${var.nvme_size_gb}Gi" } } }
        }]
      }
    }
  }
  depends_on = [helm_release.ch_operator, aws_eks_node_group.analytics]
}

resource "aws_s3_bucket" "backup" {
  bucket = var.backup_bucket
  tags   = var.tags
}
resource "aws_s3_bucket_server_side_encryption_configuration" "backup" {
  bucket = aws_s3_bucket.backup.id
  rule { apply_server_side_encryption_by_default { sse_algorithm = "aws:kms", kms_master_key_id = var.kms_key_arn } }
}
resource "aws_s3_bucket_lifecycle_configuration" "backup" {
  bucket = aws_s3_bucket.backup.id
  rule { id = "expire", status = "Enabled", expiration { days = 35 } }
}
