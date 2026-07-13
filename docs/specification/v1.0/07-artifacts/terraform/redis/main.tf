resource "aws_elasticache_subnet_group" "this" { name = var.name  subnet_ids = var.subnet_ids }
resource "aws_security_group" "redis" {
  name = "${var.name}-redis"  vpc_id = var.vpc_id
  ingress { from_port = 6379, to_port = 6379, protocol = "tcp", cidr_blocks = ["10.0.0.0/8"] }
  tags = var.tags
}
resource "aws_elasticache_replication_group" "this" {
  replication_group_id       = var.name
  description                = "nydux cache + redlock"
  engine                     = "redis"
  engine_version             = "7.1"
  node_type                  = var.node_type
  num_cache_clusters         = 2
  automatic_failover_enabled = true
  at_rest_encryption_enabled = true
  transit_encryption_enabled = true
  kms_key_id                 = var.kms_key_arn
  subnet_group_name          = aws_elasticache_subnet_group.this.name
  security_group_ids         = [aws_security_group.redis.id]
  tags = var.tags
}
