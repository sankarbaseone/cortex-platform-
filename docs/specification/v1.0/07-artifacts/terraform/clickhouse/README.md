# terraform module: clickhouse
Deploys the NYDUX analytics store: dedicated tainted NVMe node group, Altinity
operator, replicated ClickHouseInstallation (shards x replicas), S3 cold tier,
KMS-encrypted backup bucket with 35-day lifecycle.
Inputs/outputs: see variables.tf / outputs.tf. DR: nightly `BACKUP TO S3`
CronJob is installed by charts/nydux-platform (not this module). Scaling
trigger and sizing rules: ECD-006 §6.4/§6.5.
