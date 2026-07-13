variable "name"        { type = string }
variable "subnet_ids"  { type = list(string) }
variable "vpc_id"      { type = string }
variable "node_type"   { type = string default = "cache.r6g.large" }
variable "kms_key_arn" { type = string }
variable "tags"        { type = map(string) default = {} }
