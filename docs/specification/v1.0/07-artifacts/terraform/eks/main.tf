resource "aws_eks_cluster" "this" {
  name     = var.name
  version  = var.k8s_version
  role_arn = aws_iam_role.cluster.arn
  vpc_config { subnet_ids = var.subnet_ids, endpoint_private_access = true, endpoint_public_access = false }
  encryption_config { provider { key_arn = var.kms_key_arn }, resources = ["secrets"] }
  tags = var.tags
}
resource "aws_iam_role" "cluster" {
  name = "${var.name}-eks"
  assume_role_policy = jsonencode({ Version = "2012-10-17",
    Statement = [{ Effect = "Allow", Principal = { Service = "eks.amazonaws.com" }, Action = "sts:AssumeRole" }] })
}
resource "aws_iam_role_policy_attachment" "cluster" {
  role = aws_iam_role.cluster.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSClusterPolicy"
}
resource "aws_iam_role" "node" {
  name = "${var.name}-node"
  assume_role_policy = jsonencode({ Version = "2012-10-17",
    Statement = [{ Effect = "Allow", Principal = { Service = "ec2.amazonaws.com" }, Action = "sts:AssumeRole" }] })
}
resource "aws_iam_role_policy_attachment" "node" {
  for_each = toset(["AmazonEKSWorkerNodePolicy","AmazonEKS_CNI_Policy","AmazonEC2ContainerRegistryReadOnly"])
  role = aws_iam_role.node.name
  policy_arn = "arn:aws:iam::aws:policy/${each.key}"
}
resource "aws_eks_node_group" "cp" {
  cluster_name    = aws_eks_cluster.this.name
  node_group_name = "cp"
  node_role_arn   = aws_iam_role.node.arn
  subnet_ids      = var.subnet_ids
  instance_types  = [var.cp_node_type]
  scaling_config { desired_size = var.cp_min, min_size = var.cp_min, max_size = var.cp_max }
  labels = { "nydux.ai/pool" = "cp" }
  tags = var.tags
}
resource "aws_eks_node_group" "bench" {
  cluster_name    = aws_eks_cluster.this.name
  node_group_name = "bench"
  node_role_arn   = aws_iam_role.node.arn
  subnet_ids      = var.subnet_ids
  instance_types  = [var.bench_gpu_type]
  scaling_config { desired_size = 0, min_size = 0, max_size = var.bench_max }
  labels = { "nydux.ai/pool" = "bench", "nydux.ai/bench" = "true" }
  taint { key = "nvidia.com/gpu", value = "present", effect = "NO_SCHEDULE" }
  tags = var.tags
}
