-- Run against ivyvue.

SELECT 'ivyvue database is reachable' AS check_name;

SELECT rolname
FROM pg_roles
WHERE rolname IN ('ivyvue_app', 'ivyvue_migrator', 'ivyvue_admin')
ORDER BY rolname;

SELECT schema_name
FROM information_schema.schemata
WHERE schema_name IN ('public', 'project_data', 'app_private')
ORDER BY schema_name;

SELECT table_schema, table_name
FROM information_schema.tables
WHERE table_schema = 'public'
  AND table_name IN (
    'users',
    'accounts',
    'account_users',
    'teams',
    'team_memberships',
    'projects',
    'project_grants',
    'project_storage_locations',
    'app_sessions',
    'visiting_hours',
    'email_deliveries',
    'email_login_tokens',
    'ivy_workspace_sessions',
    'passkey_credentials',
    'passkey_challenges'
  )
ORDER BY table_name;
