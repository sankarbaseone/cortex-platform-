output "bucket_names" { value = { for k, b in aws_s3_bucket.b : k => b.bucket } }
