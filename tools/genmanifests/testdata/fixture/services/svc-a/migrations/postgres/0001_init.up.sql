-- fixture migration for tools/genmanifests tests - mirrors real conventions
-- (see services/tenant-svc/migrations/postgres/0001_init.up.sql) at minimal scale.
BEGIN;

CREATE TABLE widgets (
  widget_id  uuid PRIMARY KEY,
  tenant_id  uuid NOT NULL,
  name       text NOT NULL,
  status     text NOT NULL DEFAULT 'active'
             CHECK (status IN ('active','retired'))
);
CREATE INDEX widgets_tenant_idx ON widgets (tenant_id);

CREATE TABLE widget_tags (
  widget_id uuid NOT NULL REFERENCES widgets(widget_id),
  tag       text NOT NULL
);

DO $$
DECLARE t text;
BEGIN
  FOREACH t IN ARRAY ARRAY['widgets']
  LOOP
    EXECUTE format('ALTER TABLE %I ENABLE ROW LEVEL SECURITY', t);
    EXECUTE format('ALTER TABLE %I FORCE ROW LEVEL SECURITY', t);
  END LOOP;
END $$;

COMMIT;
