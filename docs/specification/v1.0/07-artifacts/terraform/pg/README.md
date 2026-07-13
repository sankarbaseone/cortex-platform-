# terraform module: pg
Multi-AZ RDS Postgres 16 (or Timescale via var.timescale=true, second
instantiation). WAL-G continuous archiving runs in-cluster (chart) against the
buckets module; RDS automated backups are belt-and-braces (35d). RLS enforced
at schema level (migrations), row_security locked on here.
