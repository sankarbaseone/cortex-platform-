resource "aws_security_group" "kafka" {
  name = "${var.name}-kafka"  vpc_id = var.vpc_id
  ingress { from_port = 9096, to_port = 9096, protocol = "tcp", cidr_blocks = ["10.0.0.0/8"] }
  egress  { from_port = 0, to_port = 0, protocol = "-1", cidr_blocks = ["0.0.0.0/0"] }
  tags = var.tags
}
resource "aws_msk_cluster" "this" {
  cluster_name           = var.name
  kafka_version          = "3.7.x"
  number_of_broker_nodes = var.brokers
  broker_node_group_info {
    instance_type   = var.broker_type
    client_subnets  = var.subnet_ids
    security_groups = [aws_security_group.kafka.id]
    storage_info { ebs_storage_info { volume_size = var.volume_gb } }
  }
  encryption_info {
    encryption_at_rest_kms_key_arn = var.kms_key_arn
    encryption_in_transit { client_broker = "TLS", in_cluster = true }
  }
  client_authentication { sasl { scram = true } }
  tags = var.tags
}
