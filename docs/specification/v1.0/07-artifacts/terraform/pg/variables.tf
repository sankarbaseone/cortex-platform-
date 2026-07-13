variable "name"          { type = string }
variable "subnet_ids"    { type = list(string) }
variable "vpc_id"        { type = string }
variable "instance"      { type = string default = "db.r6g.xlarge" }
variable "storage_gb"    { type = number default = 500 }
variable "timescale"     { type = bool   default = false }  # second instantiation for TS
variable "kms_key_arn"   { type = string }
variable "tags"          { type = map(string) default = {} }
