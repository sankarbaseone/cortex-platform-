-- migrations/postgres/0002_seed_dev.up.sql  (dev/staging profile ONLY; guarded by env in migrate wrapper)
BEGIN;
INSERT INTO tenants (tenant_id, name, plan, region) VALUES
  ('018f0000-0000-7000-8000-000000000001','nydux-internal','internal','ap-south-1'),
  ('018f0000-0000-7000-8000-000000000002','design-partner-a','design','ap-south-1');
INSERT INTO roles (role_id, tenant_id, name, permissions) VALUES
  ('018f0000-0000-7000-8000-00000000r001','018f0000-0000-7000-8000-000000000001','tenant-admin','{"*":["*"]}'),
  ('018f0000-0000-7000-8000-00000000r002','018f0000-0000-7000-8000-000000000001','viewer','{"read":["*"]}');
COMMIT;
