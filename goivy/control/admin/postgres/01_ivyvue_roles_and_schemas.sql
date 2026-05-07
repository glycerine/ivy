-- Run against the ivyvue database as the local bootstrap role.

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'ivyvue_app') THEN
    CREATE ROLE ivyvue_app LOGIN PASSWORD 'ivyvue_app_dev';
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'ivyvue_migrator') THEN
    CREATE ROLE ivyvue_migrator LOGIN PASSWORD 'ivyvue_migrator_dev';
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'ivyvue_admin') THEN
    CREATE ROLE ivyvue_admin LOGIN PASSWORD 'ivyvue_admin_dev';
  END IF;
END
$$;

REVOKE ALL ON DATABASE ivyvue FROM PUBLIC;
GRANT CONNECT ON DATABASE ivyvue TO ivyvue_app, ivyvue_migrator, ivyvue_admin;

CREATE SCHEMA IF NOT EXISTS control AUTHORIZATION ivyvue_migrator;
CREATE SCHEMA IF NOT EXISTS project_data AUTHORIZATION ivyvue_migrator;
CREATE SCHEMA IF NOT EXISTS app_private AUTHORIZATION ivyvue_migrator;

REVOKE ALL ON SCHEMA control, project_data, app_private FROM PUBLIC;
GRANT USAGE ON SCHEMA control, project_data TO ivyvue_app;
GRANT USAGE, CREATE ON SCHEMA control, project_data, app_private TO ivyvue_migrator;
GRANT USAGE, CREATE ON SCHEMA control, project_data, app_private TO ivyvue_admin;

ALTER DEFAULT PRIVILEGES FOR ROLE ivyvue_migrator IN SCHEMA control
  GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO ivyvue_app;
ALTER DEFAULT PRIVILEGES FOR ROLE ivyvue_migrator IN SCHEMA project_data
  GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO ivyvue_app;
ALTER DEFAULT PRIVILEGES FOR ROLE ivyvue_migrator IN SCHEMA control
  GRANT USAGE, SELECT ON SEQUENCES TO ivyvue_app;
ALTER DEFAULT PRIVILEGES FOR ROLE ivyvue_migrator IN SCHEMA project_data
  GRANT USAGE, SELECT ON SEQUENCES TO ivyvue_app;
