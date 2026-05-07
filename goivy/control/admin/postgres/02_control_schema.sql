-- Initial control-plane schema. Run against ivyvue as ivyvue_migrator.

CREATE TABLE IF NOT EXISTS control.schema_migrations (
  version text PRIMARY KEY,
  applied_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS control.users (
  id uuid PRIMARY KEY,
  idp_issuer text NOT NULL,
  idp_subject text NOT NULL,
  email text NOT NULL,
  display_name text NOT NULL DEFAULT '',
  email_verified_at timestamptz,
  disabled_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (idp_issuer, idp_subject),
  UNIQUE (email)
);

CREATE TABLE IF NOT EXISTS control.accounts (
  id uuid PRIMARY KEY,
  slug text NOT NULL UNIQUE,
  display_name text NOT NULL,
  billing_email text NOT NULL,
  billing_status text NOT NULL DEFAULT 'missing_payment_method',
  disabled_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS control.account_users (
  account_id uuid NOT NULL REFERENCES control.accounts(id),
  user_id uuid NOT NULL REFERENCES control.users(id),
  role text NOT NULL CHECK (role IN ('owner', 'admin', 'billing_admin', 'member')),
  seat_state text NOT NULL DEFAULT 'active',
  disabled_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (account_id, user_id)
);

CREATE TABLE IF NOT EXISTS control.teams (
  id uuid PRIMARY KEY,
  account_id uuid NOT NULL REFERENCES control.accounts(id),
  slug text NOT NULL,
  display_name text NOT NULL,
  disabled_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (account_id, slug)
);

CREATE TABLE IF NOT EXISTS control.team_memberships (
  team_id uuid NOT NULL REFERENCES control.teams(id),
  user_id uuid NOT NULL REFERENCES control.users(id),
  role text NOT NULL CHECK (role IN ('owner', 'admin', 'member')),
  disabled_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (team_id, user_id)
);

CREATE TABLE IF NOT EXISTS control.projects (
  id uuid PRIMARY KEY,
  account_id uuid NOT NULL REFERENCES control.accounts(id),
  slug text NOT NULL,
  display_name text NOT NULL,
  created_by_user_id uuid NOT NULL REFERENCES control.users(id),
  disabled_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (account_id, slug)
);

CREATE TABLE IF NOT EXISTS control.project_grants (
  project_id uuid NOT NULL REFERENCES control.projects(id),
  subject_kind text NOT NULL CHECK (subject_kind IN ('user', 'team', 'account')),
  subject_id uuid NOT NULL,
  role text NOT NULL CHECK (role IN ('read', 'write', 'admin')),
  disabled_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (project_id, subject_kind, subject_id)
);

CREATE TABLE IF NOT EXISTS control.project_storage_locations (
  project_id uuid PRIMARY KEY REFERENCES control.projects(id),
  mode text NOT NULL DEFAULT 'shared_postgres',
  database_name text NOT NULL DEFAULT 'ivyvue',
  schema_name text NOT NULL DEFAULT 'project_data',
  state text NOT NULL DEFAULT 'ready',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS control.app_sessions (
  id_hash bytea PRIMARY KEY,
  user_id uuid NOT NULL REFERENCES control.users(id),
  csrf_token_hash bytea NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  last_seen_at timestamptz NOT NULL DEFAULT now(),
  idle_expires_at timestamptz NOT NULL,
  absolute_expires_at timestamptz NOT NULL,
  revoked_at timestamptz
);

CREATE TABLE IF NOT EXISTS control.email_login_tokens (
  token_hash bytea PRIMARY KEY,
  email text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  expires_at timestamptz NOT NULL,
  used_at timestamptz
);

CREATE INDEX IF NOT EXISTS email_login_tokens_email_created_idx
  ON control.email_login_tokens (email, created_at DESC);

CREATE TABLE IF NOT EXISTS control.ivy_workspace_sessions (
  id uuid PRIMARY KEY,
  project_id uuid NOT NULL REFERENCES control.projects(id),
  user_id uuid NOT NULL REFERENCES control.users(id),
  analysis_session_id text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  last_seen_at timestamptz NOT NULL DEFAULT now(),
  closed_at timestamptz
);
