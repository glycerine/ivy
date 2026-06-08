-- Local bootstrap script. Run with a PostgreSQL role that can create databases.
-- The development bootstrap role is jaten/jaten.

SELECT 'CREATE DATABASE ivyvue'
WHERE NOT EXISTS (SELECT 1 FROM pg_database WHERE datname = 'ivyvue')
\gexec
