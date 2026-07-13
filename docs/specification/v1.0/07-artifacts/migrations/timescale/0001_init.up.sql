-- migrations/timescale/0001_init.up.sql
-- Business time-series per RFC-007 §C.3.
BEGIN;

CREATE EXTENSION IF NOT EXISTS timescaledb;

CREATE TABLE cost_slices (
  time        timestamptz NOT NULL,
  tenant_id   uuid NOT NULL,
  slice_id    uuid NOT NULL,
  cluster_id  uuid,
  team_id     uuid,
  model_id    text,
  workload_id text,
  usd_micros  bigint NOT NULL CHECK (usd_micros >= 0),
  basis       text NOT NULL CHECK (basis IN ('ondemand','commit','spot','amortized')),
  gpu_hours   double precision,
  tokens      bigint
);
SELECT create_hypertable('cost_slices','time', chunk_time_interval => interval '1 day');
CREATE INDEX cost_slices_tenant_team_idx ON cost_slices (tenant_id, team_id, time DESC);
CREATE INDEX cost_slices_tenant_model_idx ON cost_slices (tenant_id, model_id, time DESC);
ALTER TABLE cost_slices SET (timescaledb.compress,
  timescaledb.compress_segmentby = 'tenant_id,team_id',
  timescaledb.compress_orderby   = 'time DESC');
SELECT add_compression_policy('cost_slices', interval '7 days');
SELECT add_retention_policy('cost_slices',  interval '25 months');

CREATE TABLE savings_ledger (
  time                  timestamptz NOT NULL,
  tenant_id             uuid NOT NULL,
  period                text NOT NULL CHECK (period ~ '^[0-9]{4}-[0-9]{2}$'),
  baseline_id           uuid NOT NULL,
  baseline_version      int  NOT NULL,
  twin_usd_micros       bigint NOT NULL,
  trailing_usd_micros   bigint NOT NULL,
  contractual_usd_micros bigint NOT NULL,
  method_doc_uri        text NOT NULL
);
SELECT create_hypertable('savings_ledger','time', chunk_time_interval => interval '30 days');
CREATE UNIQUE INDEX savings_ledger_period_uq ON savings_ledger (tenant_id, period, time);

-- Continuous aggregates (6h refresh lag per RFC-005 B.7 watermark)
CREATE MATERIALIZED VIEW cost_daily_by_team
WITH (timescaledb.continuous) AS
SELECT time_bucket('1 day', time) AS day, tenant_id, team_id,
       sum(usd_micros) AS usd_micros, sum(gpu_hours) AS gpu_hours, sum(tokens) AS tokens
FROM cost_slices GROUP BY 1,2,3 WITH NO DATA;
SELECT add_continuous_aggregate_policy('cost_daily_by_team',
  start_offset => interval '3 days', end_offset => interval '6 hours',
  schedule_interval => interval '1 hour');

CREATE MATERIALIZED VIEW unit_cost_daily
WITH (timescaledb.continuous) AS
SELECT time_bucket('1 day', time) AS day, tenant_id, model_id,
       sum(usd_micros)::double precision / NULLIF(sum(tokens),0) AS usd_micros_per_token
FROM cost_slices GROUP BY 1,2,3 WITH NO DATA;
SELECT add_continuous_aggregate_policy('unit_cost_daily',
  start_offset => interval '3 days', end_offset => interval '6 hours',
  schedule_interval => interval '1 hour');

COMMIT;

-- ============================================================
-- migrations/clickhouse/0001_init.sql  (clickhouse-migrations; ON CLUSTER 'nydux')
-- Telemetry + kernel analytics per RFC-007 §C.4.
-- ============================================================
CREATE TABLE IF NOT EXISTS gpu_samples ON CLUSTER 'nydux' (
  ts DateTime64(3),
  tenant LowCardinality(String),
  cluster LowCardinality(String),
  gpu_uuid String,
  node LowCardinality(String),
  pod String,
  ns LowCardinality(String),
  sm_util Float32, mem_util Float32, mem_used_mb UInt32,
  power_w Float32, temp_c Float32, sm_clock UInt16,
  pcie_tx_mb Float32, nvlink_tx_mb Float32,
  tensor_active Float32, dram_active Float32,
  xid UInt16 DEFAULT 0
) ENGINE = ReplicatedMergeTree('/ch/{shard}/gpu_samples','{replica}')
PARTITION BY toYYYYMMDD(ts)
ORDER BY (tenant, cluster, gpu_uuid, ts)
TTL toDateTime(ts) + INTERVAL 14 DAY TO VOLUME 'cold',
    toDateTime(ts) + INTERVAL 90 DAY DELETE
SETTINGS index_granularity = 8192;

CREATE TABLE IF NOT EXISTS gpu_samples_1m ON CLUSTER 'nydux' (
  ts DateTime, tenant LowCardinality(String), cluster LowCardinality(String), gpu_uuid String,
  sm_util AggregateFunction(avg, Float32),
  mem_util AggregateFunction(avg, Float32),
  power_w AggregateFunction(avg, Float32),
  tensor_active AggregateFunction(avg, Float32),
  samples AggregateFunction(count)
) ENGINE = ReplicatedAggregatingMergeTree('/ch/{shard}/gpu_samples_1m','{replica}')
PARTITION BY toYYYYMM(ts)
ORDER BY (tenant, cluster, gpu_uuid, ts)
TTL ts + INTERVAL 25 MONTH DELETE;

CREATE MATERIALIZED VIEW IF NOT EXISTS gpu_samples_1m_mv ON CLUSTER 'nydux'
TO gpu_samples_1m AS
SELECT toStartOfMinute(ts) AS ts, tenant, cluster, gpu_uuid,
  avgState(sm_util) AS sm_util, avgState(mem_util) AS mem_util,
  avgState(power_w) AS power_w, avgState(tensor_active) AS tensor_active,
  countState() AS samples
FROM gpu_samples GROUP BY ts, tenant, cluster, gpu_uuid;

CREATE TABLE IF NOT EXISTS kernel_events ON CLUSTER 'nydux' (
  ts DateTime64(3),
  tenant LowCardinality(String),
  kernel_hash FixedString(64),
  family_hash FixedString(64),
  arch LowCardinality(String),
  toolchain_fp String,
  kes Float32, comp Map(String, Float32),
  status Enum8('SCORED'=1,'STATIC'=2,'MEAS_ONLY'=3),
  confidence Float32,
  dur_us Float64, launches UInt32
) ENGINE = ReplicatedMergeTree('/ch/{shard}/kernel_events','{replica}')
PARTITION BY toYYYYMM(ts)
ORDER BY (tenant, family_hash, kernel_hash, ts)
TTL toDateTime(ts) + INTERVAL 25 MONTH DELETE;

CREATE TABLE IF NOT EXISTS serving_metrics ON CLUSTER 'nydux' (
  ts DateTime64(3),
  tenant LowCardinality(String), cluster LowCardinality(String),
  deploy LowCardinality(String),
  ttft_ms Float32, tpot_ms Float32, batch UInt16,
  kv_hit Float32, req UInt32, tokens_in UInt64, tokens_out UInt64
) ENGINE = ReplicatedMergeTree('/ch/{shard}/serving_metrics','{replica}')
PARTITION BY toYYYYMMDD(ts)
ORDER BY (tenant, cluster, deploy, ts)
TTL toDateTime(ts) + INTERVAL 90 DAY DELETE;

-- Row policies (per RFC-007 C.4 / OQ-12): one per tenant, managed by tenant-svc:
-- CREATE ROW POLICY rp_<tenant> ON gpu_samples FOR SELECT USING tenant = '<tenant>' TO nydux_reader_<tenant>;
