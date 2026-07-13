-- migrations/postgres/0001_init.up.sql
-- Control-plane transactional schema. Implements RFC-007 §C.2 verbatim + ECD constraints.
-- Every tenant table: RLS enabled, policy t_iso. App role: nydux_app (no BYPASSRLS).

BEGIN;

CREATE EXTENSION IF NOT EXISTS citext;
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE tenants (
  tenant_id   uuid PRIMARY KEY,
  name        text NOT NULL,
  plan        text NOT NULL,
  region      text NOT NULL,
  created_at  timestamptz NOT NULL DEFAULT now(),
  status      text NOT NULL DEFAULT 'active'
              CHECK (status IN ('active','suspended','offboarding'))
);

CREATE TABLE clusters (
  cluster_id   uuid PRIMARY KEY,
  tenant_id    uuid NOT NULL REFERENCES tenants(tenant_id),
  display_name text,
  provider     text,
  topology     jsonb,
  created_at   timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX clusters_tenant_idx ON clusters (tenant_id);

CREATE TABLE users (
  user_id   uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(tenant_id),
  email     citext NOT NULL,
  idp_sub   text NOT NULL,
  UNIQUE (tenant_id, email)
);

CREATE TABLE roles (
  role_id     uuid PRIMARY KEY,
  tenant_id   uuid REFERENCES tenants(tenant_id),
  name        text NOT NULL,
  permissions jsonb NOT NULL DEFAULT '{}'
);

CREATE TABLE user_roles (
  user_id uuid NOT NULL REFERENCES users(user_id),
  role_id uuid NOT NULL REFERENCES roles(role_id),
  scope   jsonb NOT NULL DEFAULT '{}',
  PRIMARY KEY (user_id, role_id)
);

CREATE TABLE api_keys (
  key_id     uuid PRIMARY KEY,
  tenant_id  uuid NOT NULL REFERENCES tenants(tenant_id),
  hash       bytea NOT NULL,
  prefix     text  NOT NULL,
  scopes     text[] NOT NULL,
  created_by uuid,
  expires_at timestamptz,
  revoked_at timestamptz
);
CREATE UNIQUE INDEX api_keys_prefix_uq ON api_keys (prefix);

CREATE TABLE toolchains (
  toolchain_id uuid PRIMARY KEY,
  tenant_id    uuid NOT NULL REFERENCES tenants(tenant_id),
  name         text NOT NULL,
  cuda_ver     text, ptxas_ver text, triton_ver text, torch_ver text, host_cc text,
  fingerprint  text NOT NULL CHECK (fingerprint ~ '^[a-f0-9]{64}$'),
  approval     text NOT NULL DEFAULT 'unreviewed'
               CHECK (approval IN ('unreviewed','approved','revoked')),
  approved_by  uuid,
  approved_at  timestamptz,
  UNIQUE (tenant_id, fingerprint)
);

CREATE TABLE kernels (
  kernel_hash       text PRIMARY KEY CHECK (kernel_hash ~ '^[a-f0-9]{64}$'),
  tenant_id         uuid NOT NULL REFERENCES tenants(tenant_id),
  family_hash       text NOT NULL CHECK (family_hash ~ '^[a-f0-9]{64}$'),
  arch              text NOT NULL,
  status            text NOT NULL
                    CHECK (status IN ('SCORED','STATIC_ESTIMATE','MEASUREMENT_ONLY')),
  kes               numeric(5,2) CHECK (kes >= 0 AND kes <= 100),
  kes_model_version text,
  confidence        numeric(3,2) CHECK (confidence >= 0 AND confidence <= 1),
  toolchain_fp      text,
  first_seen        timestamptz NOT NULL,
  last_seen         timestamptz NOT NULL,
  CONSTRAINT scored_requires_score CHECK (status <> 'SCORED' OR kes IS NOT NULL)
);
CREATE INDEX kernels_tenant_family_idx ON kernels (tenant_id, family_hash);
CREATE INDEX kernels_tenant_kes_idx ON kernels (tenant_id, kes) WHERE status = 'SCORED';
CREATE INDEX kernels_tenant_lastseen_idx ON kernels (tenant_id, last_seen DESC);

CREATE TABLE recommendations (
  rec_id      uuid PRIMARY KEY,
  tenant_id   uuid NOT NULL REFERENCES tenants(tenant_id),
  kernel_hash text REFERENCES kernels(kernel_hash),
  pattern_id  text NOT NULL CHECK (pattern_id IN (
    'unfused_elementwise','missing_flashattn','reg_pressure','noncoalesced',
    'fp32_gemm_tc','smallbatch_gemm','launch_storm','triton_tile',
    'cublas_fallback','comm_overlap_gap','toolchain_upgrade','toolchain_holdback')),
  gain_p50    numeric, gain_p90 numeric,
  confidence  numeric(3,2) CHECK (confidence >= 0 AND confidence <= 1),
  effort      text CHECK (effort IN ('trivial','low','medium','high')),
  risk        numeric CHECK (risk >= 0),
  state       text NOT NULL DEFAULT 'created' CHECK (state IN
    ('created','approved','rejected','applied','verified','failed','rolled_back')),
  evidence    jsonb NOT NULL,
  patch_uri   text,
  created_at  timestamptz NOT NULL DEFAULT now(),
  decided_by  uuid,
  decided_at  timestamptz,
  rationale   text
);
CREATE INDEX recs_tenant_state_idx ON recommendations (tenant_id, state, created_at DESC);

CREATE TABLE policies (
  policy_id   uuid PRIMARY KEY,
  tenant_id   uuid NOT NULL REFERENCES tenants(tenant_id),
  name        text NOT NULL,
  rego        text NOT NULL,
  enforcement text NOT NULL CHECK (enforcement IN ('block','warn','audit')),
  version     int  NOT NULL DEFAULT 1,
  updated_at  timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE audit_entries (
  tenant_id  uuid   NOT NULL,
  seq        bigint NOT NULL,
  entry_hash bytea  NOT NULL CHECK (octet_length(entry_hash) = 32),
  prev_hash  bytea  NOT NULL CHECK (octet_length(prev_hash)  = 32),
  actor      text   NOT NULL,
  action     text   NOT NULL,
  subject    jsonb  NOT NULL,
  at         timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (tenant_id, seq)
);
-- No UPDATE/DELETE grants ever issued on audit_entries (append-only by grant, RFC-009 I.8).

CREATE TABLE savings_baselines (
  baseline_id        uuid PRIMARY KEY,
  tenant_id          uuid NOT NULL REFERENCES tenants(tenant_id),
  version            int  NOT NULL,
  frozen_state       jsonb NOT NULL,
  frozen_state_hash  text NOT NULL CHECK (frozen_state_hash ~ '^[a-f0-9]{64}$'),
  twin_model_version text NOT NULL,
  anchored_at        timestamptz NOT NULL,
  cosign_customer    uuid,
  cosign_nydux       uuid,
  reanchor_reason    text,
  UNIQUE (tenant_id, version)
);

CREATE TABLE jobs (
  job_id      uuid PRIMARY KEY,
  tenant_id   uuid REFERENCES tenants(tenant_id),
  kind        text NOT NULL,
  state       text NOT NULL DEFAULT 'queued'
              CHECK (state IN ('queued','running','done','failed')),
  request     jsonb,
  result_uri  text,
  created_at  timestamptz NOT NULL DEFAULT now(),
  finished_at timestamptz
);

-- Outbox (per RFC-012 O.4; one table per owning service schema in prod, single here for init)
CREATE TABLE outbox (
  id            bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  tenant_id     uuid  NOT NULL,
  topic         text  NOT NULL,
  partition_key text  NOT NULL,
  envelope      bytea NOT NULL,
  created_at    timestamptz NOT NULL DEFAULT now(),
  published_at  timestamptz
);
CREATE INDEX outbox_unpublished_idx ON outbox (id) WHERE published_at IS NULL;

-- ---------- RLS ----------
DO $$
DECLARE t text;
BEGIN
  FOREACH t IN ARRAY ARRAY['clusters','users','roles','user_roles','api_keys','toolchains',
    'kernels','recommendations','policies','audit_entries','savings_baselines','jobs','outbox']
  LOOP
    EXECUTE format('ALTER TABLE %I ENABLE ROW LEVEL SECURITY', t);
    EXECUTE format('ALTER TABLE %I FORCE ROW LEVEL SECURITY', t);
    EXECUTE format(
      'CREATE POLICY t_iso ON %I USING (tenant_id = current_setting(''nydux.tenant'')::uuid)', t);
  END LOOP;
END $$;

COMMIT;
