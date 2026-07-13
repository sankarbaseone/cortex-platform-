variable "cluster_name"     { type = string  description = "EKS cluster name to deploy into" }
variable "namespace"        { type = string  default = "nydux-data" }
variable "shards"           { type = number  default = 1 }
variable "replicas"         { type = number  default = 2  description = "replicas per shard (RFC-001 A.6)" }
variable "nvme_size_gb"     { type = number  default = 2000 description = "per node (ECD-006 sizing)" }
variable "instance_type"    { type = string  default = "i4i.2xlarge" }
variable "keeper_replicas"  { type = number  default = 3 }
variable "backup_bucket"    { type = string  description = "S3 bucket for nightly incremental BACKUP TO S3" }
variable "cold_bucket"      { type = string  description = "S3 bucket backing the cold volume TTL tier" }
variable "kms_key_arn"      { type = string }
variable "tags"             { type = map(string) default = {} }
