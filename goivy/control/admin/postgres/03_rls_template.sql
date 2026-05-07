-- Template for future project_data tables.
-- Keep this file as executable documentation until the first project_data table exists.

CREATE OR REPLACE FUNCTION app_private.current_project_id()
RETURNS uuid
LANGUAGE sql
STABLE
AS $$
  SELECT NULLIF(current_setting('ivy.project_id', true), '')::uuid
$$;

CREATE OR REPLACE FUNCTION app_private.current_project_role()
RETURNS text
LANGUAGE sql
STABLE
AS $$
  SELECT NULLIF(current_setting('ivy.project_role', true), '')
$$;

-- Example for later tables:
--
-- ALTER TABLE project_data.some_table ENABLE ROW LEVEL SECURITY;
-- ALTER TABLE project_data.some_table FORCE ROW LEVEL SECURITY;
--
-- CREATE POLICY some_table_read_project
--   ON project_data.some_table
--   FOR SELECT
--   USING (project_id = app_private.current_project_id());
--
-- CREATE POLICY some_table_write_project
--   ON project_data.some_table
--   FOR INSERT
--   WITH CHECK (
--     project_id = app_private.current_project_id()
--     AND app_private.current_project_role() IN ('write', 'admin')
--   );
