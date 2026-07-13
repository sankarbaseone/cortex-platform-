-- migrations/postgres/0001_init.down.sql
BEGIN;
DROP TABLE IF EXISTS outbox, jobs, savings_baselines, audit_entries, policies,
  recommendations, kernels, toolchains, api_keys, user_roles, roles, users,
  clusters, tenants CASCADE;
COMMIT;
