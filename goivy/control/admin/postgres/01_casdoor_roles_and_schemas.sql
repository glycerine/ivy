-- Run against the casdoor database as the local bootstrap role.
-- Casdoor owns this database and manages its own application tables.

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'casdoor_app') THEN
    CREATE ROLE casdoor_app LOGIN PASSWORD 'casdoor_app_dev';
  END IF;
END
$$;

REVOKE ALL ON DATABASE casdoor FROM PUBLIC;
GRANT CONNECT, CREATE ON DATABASE casdoor TO casdoor_app;

REVOKE ALL ON SCHEMA public FROM PUBLIC;
GRANT USAGE, CREATE ON SCHEMA public TO casdoor_app;
