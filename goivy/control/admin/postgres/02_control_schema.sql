-- Initial control-plane schema. Run against ivyvue as ivyvue_migrator.

CREATE TABLE IF NOT EXISTS schema_migrations (
  version text PRIMARY KEY,
  applied_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS users (
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

CREATE TABLE IF NOT EXISTS accounts (
  id uuid PRIMARY KEY,
  slug text NOT NULL UNIQUE,
  display_name text NOT NULL,
  billing_email text NOT NULL,
  billing_status text NOT NULL DEFAULT 'missing_payment_method',
  disabled_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS account_users (
  account_id uuid NOT NULL REFERENCES accounts(id),
  user_id uuid NOT NULL REFERENCES users(id),
  role text NOT NULL CHECK (role IN ('owner', 'admin', 'billing_admin', 'member')),
  seat_state text NOT NULL DEFAULT 'active',
  disabled_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (account_id, user_id)
);

CREATE TABLE IF NOT EXISTS teams (
  id uuid PRIMARY KEY,
  account_id uuid NOT NULL REFERENCES accounts(id),
  slug text NOT NULL,
  display_name text NOT NULL,
  disabled_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (account_id, slug)
);

CREATE TABLE IF NOT EXISTS team_memberships (
  team_id uuid NOT NULL REFERENCES teams(id),
  user_id uuid NOT NULL REFERENCES users(id),
  role text NOT NULL CHECK (role IN ('owner', 'admin', 'member')),
  disabled_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (team_id, user_id)
);

CREATE TABLE IF NOT EXISTS projects (
  id uuid PRIMARY KEY,
  account_id uuid NOT NULL REFERENCES accounts(id),
  slug text NOT NULL,
  display_name text NOT NULL,
  created_by_user_id uuid NOT NULL REFERENCES users(id),
  disabled_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (account_id, slug)
);

CREATE TABLE IF NOT EXISTS project_grants (
  project_id uuid NOT NULL REFERENCES projects(id),
  subject_kind text NOT NULL CHECK (subject_kind IN ('user', 'team', 'account')),
  subject_id uuid NOT NULL,
  role text NOT NULL CHECK (role IN ('read', 'write', 'admin')),
  disabled_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (project_id, subject_kind, subject_id)
);

CREATE TABLE IF NOT EXISTS project_storage_locations (
  project_id uuid PRIMARY KEY REFERENCES projects(id),
  mode text NOT NULL DEFAULT 'shared_postgres',
  database_name text NOT NULL DEFAULT 'ivyvue',
  schema_name text NOT NULL DEFAULT 'project_data',
  state text NOT NULL DEFAULT 'ready',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS app_sessions (
  id_hash bytea PRIMARY KEY,
  user_id uuid NOT NULL REFERENCES users(id),
  csrf_token_hash bytea NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  last_seen_at timestamptz NOT NULL DEFAULT now(),
  idle_expires_at timestamptz NOT NULL,
  absolute_expires_at timestamptz NOT NULL,
  revoked_at timestamptz
);

CREATE TABLE IF NOT EXISTS visiting_hours (
  id uuid PRIMARY KEY,
  user_id uuid NOT NULL REFERENCES users(id),
  session_id_hash bytea NOT NULL REFERENCES app_sessions(id_hash),
  visited_at timestamptz NOT NULL,
  visited_hour timestamptz NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (user_id, visited_hour)
);

CREATE INDEX IF NOT EXISTS visiting_hours_user_visited_at_idx
  ON visiting_hours (user_id, visited_at DESC);

CREATE TABLE IF NOT EXISTS email_login_tokens (
  token_hash bytea PRIMARY KEY,
  email text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  expires_at timestamptz NOT NULL,
  used_at timestamptz
);

CREATE INDEX IF NOT EXISTS email_login_tokens_email_created_idx
  ON email_login_tokens (email, created_at DESC);

CREATE TABLE IF NOT EXISTS email_deliveries (
  id uuid PRIMARY KEY,
  to_email text NOT NULL,
  kind text NOT NULL DEFAULT 'login_link',
  login_url text NOT NULL,
  token_hash bytea,
  provider text NOT NULL DEFAULT 'database',
  created_at timestamptz NOT NULL DEFAULT now(),
  expires_at timestamptz NOT NULL,
  delivered_at timestamptz
);

CREATE INDEX IF NOT EXISTS email_deliveries_to_email_created_idx
  ON email_deliveries (to_email, created_at DESC);

CREATE INDEX IF NOT EXISTS email_deliveries_token_hash_idx
  ON email_deliveries (token_hash);

CREATE TABLE IF NOT EXISTS ivy_workspace_sessions (
  id uuid PRIMARY KEY,
  project_id uuid NOT NULL REFERENCES projects(id),
  user_id uuid NOT NULL REFERENCES users(id),
  analysis_session_id text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  last_seen_at timestamptz NOT NULL DEFAULT now(),
  closed_at timestamptz
);

CREATE TABLE IF NOT EXISTS passkey_credentials (
  id uuid PRIMARY KEY,
  user_id uuid NOT NULL REFERENCES users(id),
  credential_id bytea NOT NULL UNIQUE,
  public_key_cose bytea NOT NULL,
  sign_count bigint NOT NULL DEFAULT 0,
  transports text[] NOT NULL DEFAULT '{}',
  backup_eligible boolean NOT NULL DEFAULT false,
  backed_up boolean NOT NULL DEFAULT false,
  attestation_type text NOT NULL DEFAULT '',
  aaguid uuid,
  display_name text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL DEFAULT now(),
  last_used_at timestamptz,
  disabled_at timestamptz
);

CREATE INDEX IF NOT EXISTS passkey_credentials_user_idx
  ON passkey_credentials (user_id, created_at DESC);
