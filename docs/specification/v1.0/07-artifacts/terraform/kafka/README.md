# terraform module: kafka
MSK 3-broker TLS+SCRAM cluster (RF=3 per topics.yaml defaults). Topic creation
is NOT Terraform — `nyduxctl topic-apply` owns topics from kafka/topics.yaml.
Redpanda drop-in remains permitted in dev/self-host (OQ-04).
