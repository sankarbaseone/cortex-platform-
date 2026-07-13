resource "aws_db_subnet_group" "this" { name = var.name  subnet_ids = var.subnet_ids }
resource "aws_security_group" "pg" {
  name = "${var.name}-pg"  vpc_id = var.vpc_id
  ingress { from_port = 5432, to_port = 5432, protocol = "tcp", cidr_blocks = ["10.0.0.0/8"] }
  tags = var.tags
}
resource "aws_db_instance" "this" {
  identifier              = var.name
  engine                  = "postgres"
  engine_version          = "16.4"
  instance_class          = var.instance
  allocated_storage       = var.storage_gb
  max_allocated_storage   = var.storage_gb * 3   # autogrow, alert at 70% (ECD-006)
  storage_type            = "gp3"
  db_subnet_group_name    = aws_db_subnet_group.this.name
  vpc_security_group_ids  = [aws_security_group.pg.id]
  storage_encrypted       = true
  kms_key_id              = var.kms_key_arn
  multi_az                = true
  backup_retention_period = 35
  backup_window           = "20:00-21:00"
  deletion_protection     = true
  performance_insights_enabled = true
  parameter_group_name    = aws_db_parameter_group.this.name
  tags = var.tags
}
resource "aws_db_parameter_group" "this" {
  name   = "${var.name}-params"
  family = "postgres16"
  parameter { name = "archive_timeout", value = "60" }         # RPO<=5min
  parameter { name = "shared_preload_libraries", value = var.timescale ? "timescaledb,pg_stat_statements" : "pg_stat_statements", apply_method = "pending-reboot" }
  parameter { name = "row_security", value = "1" }
}
