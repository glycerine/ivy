-- Run against ivyvue.

SELECT 'ivyvue database is reachable' AS check_name;

SELECT rolname
FROM pg_roles
WHERE rolname IN ('ivyvue_app', 'ivyvue_migrator', 'ivyvue_admin')
ORDER BY rolname;

SELECT schema_name
FROM information_schema.schemata
WHERE schema_name IN ('control', 'project_data', 'app_private')
ORDER BY schema_name;
