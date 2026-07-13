variable "name"            { type = string }
variable "vpc_id"          { type = string }
variable "subnet_ids"      { type = list(string) }
variable "k8s_version"     { type = string  default = "1.30" }
variable "cp_node_type"    { type = string  default = "m6i.2xlarge" }
variable "cp_min"          { type = number  default = 3 }
variable "cp_max"          { type = number  default = 12 }
variable "bench_gpu_type"  { type = string  default = "g6.2xlarge" }
variable "bench_max"       { type = number  default = 4 }
variable "kms_key_arn"     { type = string }
variable "tags"            { type = map(string) default = {} }
