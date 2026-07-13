variable "name"         { type = string }
variable "vpc_id"       { type = string }
variable "subnet_ids"   { type = list(string) }
variable "broker_type"  { type = string default = "kafka.m5.xlarge" }
variable "brokers"      { type = number default = 3 }
variable "volume_gb"    { type = number default = 1000 }
variable "kms_key_arn"  { type = string }
variable "tags"         { type = map(string) default = {} }
