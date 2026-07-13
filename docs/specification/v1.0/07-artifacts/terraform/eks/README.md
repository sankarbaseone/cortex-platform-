# terraform module: eks
Private EKS with KMS secret encryption, `cp` general pool, and scale-to-zero
GPU `bench` pool (tainted nvidia.com/gpu) for bench-runner jobs. Analytics NVMe
pool lives in the clickhouse module. Conventions per shipped clickhouse module.
