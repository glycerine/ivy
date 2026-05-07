# Ivy Web Architecture With Auth

## Current Auth Decision

The earlier Casdoor/ZITADEL-style identity-provider plan is superseded.
Casdoor's Go backend still requires a large React/Node frontend build, which is
not acceptable for this project.

The control-plane server now owns the primary auth flow:

```text
Browser
  -> Go control-plane server
       -> PostgreSQL ivyvue database
       -> Mailgun API for production email delivery
       -> ivyvue analysis server
```

Phase 0 auth policy:

- every account starts with a verified email address
- email verification uses a 10-minute, single-use magic link
- tests use an in-memory `EmailSender`, never Mailgun
- production/manual testing uses Mailgun only when an API key is explicitly set
- successful login creates the HttpOnly `ivy_webvue_session` cookie
- app sessions use a sliding 72-hour expiry
- returning within 72 hours refreshes the cookie/session for another 72 hours
- returning after expiry requires a fresh email magic link
- OAuth sign-in through Google/GitHub can be added as an attached identity after
  the user has verified an email address
- passkeys should be offered after the initial email signup roundtrip

Implementation package shape:

```text
goivy/control/
  EmailSender interface
  MemoryEmailSender for tests
  MailgunEmailSender for explicit real delivery
  email login token store
  app session store
  optional OIDC/OAuth client adapters later
```

The large Casdoor sections below are retained only as historical context until
this document is fully rewritten around the current email-first architecture.

## Summary

The browser-facing application should be split into three processes:

```text
Browser
  -> Go control-plane server
       -> Casdoor identity provider
       -> PostgreSQL ivyvue database
       -> ivyvue analysis server
```

The Go control-plane server is the primary server for browser traffic. It serves
the Vue application, owns product authorization, owns billing/account/team/
project APIs, and brokers analysis requests to the existing Ivy/Z3 server.

Casdoor owns identity: email-based users, login, invite flows, password reset,
email verification, MFA/passkeys later, and OAuth/OIDC issuer behavior.

ZITADEL is excluded from this architecture because current ZITADEL server code
is AGPL-3.0. A commercial license could change that, but the open-source AGPL
path is not acceptable for this project.

The ivyvue analysis server owns Ivy execution and Z3. For now it keeps the
existing embedded in-process Z3 via CGO because that is the path that works
today.

## Process Responsibilities

### 1. Go Control-Plane Server

The control-plane server is the only application entrypoint users should type
into a browser.

Responsibilities:

- serve the Vue 3 web app and static assets
- initiate and complete login through Casdoor
- create and manage the local `ivy_webvue_session` cookie
- map Casdoor/OIDC users to local product users
- expose account/team/project APIs
- enforce billing account, team, and project authorization
- own PostgreSQL RLS context setup for project-scoped data access
- call the ivyvue analysis server for Ivy/Z3 work
- hold no Z3 state directly
- expose no Casdoor admin API to browsers

The control-plane server answers:

```text
Who is logged into this browser session?
What billing accounts, teams, and projects can this user access?
Is this user allowed to perform this project action?
Which analysis session/job should receive this request?
```

Suggested package/command shape:

```text
goivy/cmd/ivy-control/
  main.go

goivy/control/
  server.go
  auth/
  accounts/
  teams/
  projects/
  sessions/
  analysisclient/
  postgres/
  rls/
```

The final names can change, but the boundary should remain clear: this process
is the product/control plane, not the analysis engine.

### 2. Casdoor Identity Provider

Casdoor is the selected standalone identity provider. It is written in Go and
licensed under Apache-2.0, which fits the project requirements better than
ZITADEL's current AGPL server license.

Responsibilities:

- email/password user registration or invited setup
- email verification
- password reset
- hosted login UI
- OIDC issuer endpoints
- user identity lifecycle
- optional MFA/passkeys later
- optional external identity providers later
- service account/API access for administrative provisioning

Casdoor answers:

```text
Who is this human?
Has this email/user authenticated?
What is the stable issuer + subject identity?
```

Casdoor should have its own persistence, separate from `ivyvue`:

```text
casdoor
  Casdoor-owned PostgreSQL database

ivyvue
  application/product PostgreSQL database
```

Do not mix Casdoor's schema into the application schema. Treat Casdoor as an
identity service with public OIDC endpoints and private management/admin APIs.

Fallbacks if the Casdoor spike fails:

```text
Casdoor
  selected first spike

Ory Kratos + Ory Hydra
  Go, Apache-2.0, strong architecture, but two identity services and more
  integration work.

Keycloak
  Apache-2.0, very mature and complete, but Java and operationally heavier.
```

Decision: proceed with Casdoor unless the spike uncovers a blocking issue in
licensing, self-hosting, OIDC compatibility, user management APIs, or email
flows.

### 3. ivyvue Analysis Server

The ivyvue analysis server owns Ivy execution.

Responsibilities:

- parse/load Ivy models
- maintain analysis sessions
- run check induction and other analysis actions
- call Z3 through the existing CGO integration
- produce ARG/concept/state/details payloads
- stream analysis events/results
- cancel or time-limit work where possible

The analysis server answers:

```text
Given an authorized analysis request, what does Ivy/Z3 compute?
```

It should not own:

- browser login
- Casdoor/OIDC callbacks
- billing accounts
- team membership
- project grants
- marketing/tester onboarding

This keeps the current working Z3 path intact while leaving room for later
analysis engines:

```text
server_worker_engine
browser_wasm_engine
wanix_engine
remote_paid_worker_pool
```

## Primary Browser Flow

The initial request goes to the Go control-plane server.

```text
1. Browser opens https://app.example.com/
2. Control-plane serves Vue app.
3. Vue calls GET /auth/me.
4. If no app session exists, Vue shows login/start state.
5. User clicks login.
6. Browser goes to control-plane /auth/login.
7. Control-plane redirects to Casdoor's OIDC authorize endpoint.
8. Casdoor authenticates user.
9. Casdoor redirects back to control-plane /auth/callback.
10. Control-plane validates OIDC response.
11. Control-plane maps issuer+subject to control.users.
12. Control-plane creates ivy_webvue_session.
13. Browser returns to Vue workspace/project picker.
```

Casdoor is browser-reachable for login redirects and hosted login UI, but the
application's primary URL is the control-plane server.

## Identity Mapping

Casdoor owns identities. `ivyvue` stores local product users mapped to OIDC
subjects.

```text
control.users
  id
  idp_issuer
  idp_subject
  email
  display_name
  email_verified_at
  disabled_at
  created_at
  updated_at
```

Identity key:

```text
(idp_issuer, idp_subject)
```

Do not use email as the stable identity key. Email is important for contact,
marketing, recovery, and display, but it can change. The OIDC `iss` + `sub`
values are the durable identity.

Email can still be used for:

- invitation workflows
- alpha/beta tester assignment
- marketing exports
- initial user lookup before OIDC subject exists

## Product Authorization Model

The product model separates billing responsibility from collaboration.

```text
User
  login identity mapped from Casdoor/OIDC

Account
  billing/payment responsibility
  pays for users, teams, and projects

AccountUser
  user covered by or administering a billing account

Team
  collaboration group inside an account

TeamMembership
  user membership in a team

Project
  Ivy workspace/data container paid for by one account

ProjectGrant
  read/write/admin access for a user, team, or all account users
```

An account can pay for:

- standalone users
- multiple teams
- multiple projects

A user can be:

- a standalone account user
- a member of one or more teams
- both

Project access can be granted to:

- an individual user
- a team
- all active users in the billing account

Phase 0 should keep grants within the project billing account. Cross-account
project sharing can come later.

## PostgreSQL Application Schema

Use PostgreSQL through Go `database/sql`. Prefer the pgx stdlib driver:

```text
github.com/jackc/pgx/v5/stdlib
```

Use one application database:

```text
ivyvue
```

Suggested schemas:

```text
control
  users, accounts, teams, projects, grants, app sessions, migrations

project_data
  project-owned Ivy data protected by RLS

app_private
  helper functions not callable by the runtime app role
```

Core control tables:

```text
control.users
  id
  idp_issuer
  idp_subject
  email
  display_name
  email_verified_at
  disabled_at
  created_at
  updated_at

control.accounts
  id
  slug
  display_name
  billing_email
  billing_status
  disabled_at
  created_at
  updated_at

control.account_users
  account_id
  user_id
  role              -- owner | admin | billing_admin | member
  seat_state
  disabled_at
  created_at
  updated_at

control.teams
  id
  account_id
  slug
  display_name
  disabled_at
  created_at
  updated_at

control.team_memberships
  team_id
  user_id
  role              -- owner | admin | member
  disabled_at
  created_at
  updated_at

control.projects
  id
  account_id
  slug
  display_name
  created_by_user_id
  disabled_at
  created_at
  updated_at

control.project_grants
  project_id
  subject_kind      -- user | team | account
  subject_id
  role              -- read | write | admin
  disabled_at
  created_at
  updated_at

control.project_storage_locations
  project_id
  mode              -- shared_postgres for Phase 0
  database_name
  schema_name
  state
  created_at
  updated_at

control.app_sessions
  id_hash
  user_id
  csrf_token_hash
  created_at
  last_seen_at
  idle_expires_at
  absolute_expires_at
  revoked_at

control.ivy_workspace_sessions
  id
  project_id
  user_id
  analysis_session_id
  created_at
  last_seen_at
  closed_at
```

## Row-Level Security

Project-owned data tables should include `project_id` and use PostgreSQL RLS.

Project-scoped requests run in a transaction. After application authorization
has selected an effective project role, the control-plane server sets
transaction-local settings:

```text
ivy.user_id
ivy.project_id
ivy.project_role
```

Use transaction-local `set_config`:

```sql
select set_config('ivy.user_id', $1, true);
select set_config('ivy.project_id', $2, true);
select set_config('ivy.project_role', $3, true);
```

RLS policies read these settings:

```sql
current_setting('ivy.project_id', true)
current_setting('ivy.project_role', true)
```

Every future `project_data` table must:

- include `project_id`
- enable RLS
- force RLS
- define read `USING` policies
- define write `WITH CHECK` policies

## Control-Plane To Analysis Server Boundary

The browser should not call the analysis server directly.

```text
Browser -> Control-plane -> Analysis server
```

The control-plane server:

1. validates the app session
2. checks project access
3. creates or resolves a workspace session
4. sends an internal request to the analysis server
5. returns normalized results to the browser

Internal analysis requests should include a short-lived internal credential,
not the browser session cookie.

Possible internal credential:

```text
signed analysis capability
  project_id
  user_id
  workspace_session_id
  allowed_actions
  expires_at
```

The analysis server verifies the signature and expiration. It does not need to
query billing/team/project authorization for every request.

Initial internal API shape:

```text
POST /internal/analysis/session/new
POST /internal/analysis/session/{id}/load
GET  /internal/analysis/session/{id}/arg
GET  /internal/analysis/session/{id}/concept
GET  /internal/analysis/session/{id}/menus
POST /internal/analysis/session/{id}/check
POST /internal/analysis/session/{id}/action
POST /internal/analysis/session/{id}/arg/action
GET  /internal/analysis/session/{id}/events
POST /internal/analysis/session/{id}/cancel
```

The public control-plane API can remain project-scoped:

```text
GET  /auth/me
POST /auth/login
POST /auth/logout

GET  /api/accounts
GET  /api/accounts/{accountID}/users
GET  /api/accounts/{accountID}/teams
GET  /api/accounts/{accountID}/projects
GET  /api/projects

POST /api/projects/{projectID}/ivy/session/new
POST /api/projects/{projectID}/ivy/session/{workspaceSessionID}/load
GET  /api/projects/{projectID}/ivy/session/{workspaceSessionID}/arg
GET  /api/projects/{projectID}/ivy/session/{workspaceSessionID}/concept
GET  /api/projects/{projectID}/ivy/session/{workspaceSessionID}/menus
POST /api/projects/{projectID}/ivy/session/{workspaceSessionID}/check
POST /api/projects/{projectID}/ivy/session/{workspaceSessionID}/action
POST /api/projects/{projectID}/ivy/session/{workspaceSessionID}/arg/action
GET  /api/projects/{projectID}/ivy/session/{workspaceSessionID}/events
POST /api/projects/{projectID}/ivy/session/{workspaceSessionID}/cancel
```

## Sessions And Cookies

The browser should receive only the control-plane application session cookie.

```text
ivy_webvue_session
```

Cookie properties:

- `HttpOnly`
- `SameSite=Lax`
- `Secure` in HTTPS environments
- `Path=/`
- no broad `Domain` unless explicitly needed

The Vue app should not store auth tokens in localStorage or sessionStorage.

Identity-provider tokens should be handled server-side by the control-plane
callback flow. The browser app should deal with the control-plane app session,
not raw OIDC tokens.

## Identity Provider Integration

Casdoor should own:

- user registration
- email verification
- password reset
- login UI
- user lock/deactivate
- MFA/passkeys later
- invite/setup flows for alpha and beta testers

The control-plane server should own:

- mapping Casdoor subjects to local users
- alpha/beta access grants
- billing account membership
- team membership
- project grants
- app-specific sessions

For initial testers:

```text
control admin creates/invites user in Casdoor
control admin creates local control.users row
control admin creates account/team/project rows
control admin grants project access
tester logs in through Casdoor
```

The admin path may be a Go command:

```text
ivy-control-admin invite-user
ivy-control-admin create-account
ivy-control-admin add-account-user
ivy-control-admin create-team
ivy-control-admin add-team-user
ivy-control-admin create-project
ivy-control-admin grant-project
ivy-control-admin list-users
ivy-control-admin list-accounts
ivy-control-admin list-teams
ivy-control-admin list-projects
```

This command can call Casdoor's management APIs and update `ivyvue`.

## Casdoor Configuration Plan

Use Casdoor as a separate process with its own PostgreSQL database.

Initial local naming:

```text
Casdoor endpoint:     http://127.0.0.1:18082
Casdoor database:     casdoor
Casdoor organization: ivy
Casdoor application:  ivy-control-local
OIDC client id:       ivy-control-local
Redirect URI:         http://127.0.0.1:18080/auth/callback
Logout redirect URI:  http://127.0.0.1:18080/
```

Control-plane OIDC configuration:

```text
issuer_url
client_id
client_secret
redirect_url
scopes: openid profile email
```

The control-plane server should use Authorization Code Flow with PKCE where
supported. The browser receives only the control-plane app session cookie, not
Casdoor tokens.

Casdoor should be configured to support:

- email/password login
- email verification
- password reset
- admin-created or invited users
- user disabled/locked state
- future MFA/passkeys if needed

The first spike should verify Casdoor exposes enough API surface for:

- create or invite a user by email
- mark or require email verification
- disable/deactivate a user
- query user by ID/email
- update display name/email metadata needed by `control.users`
- create an OIDC application/client
- configure redirect URLs
- export enough stable identifiers for `idp_issuer + idp_subject`

Keep Casdoor roles/groups out of the product authorization source of truth for
Phase 0. Casdoor may know identities, but `ivyvue` owns billing accounts,
teams, projects, and grants.

Casdoor references used for this plan:

- Try with Docker: https://casdoor.org/docs/basic/try-with-docker/
- Application configuration: https://casdoor.org/docs/application/config/
- OAuth 2.0 authorization code flow:
  https://casdoor.org/docs/how-to-connect/oauth/
- SDK overview: https://casdoor.org/docs/how-to-connect/sdk
- Public API overview: https://casdoor.org/docs/basic/public-api
- Invitation registration:
  https://casdoor.ai/docs/invitation/overview

## Databases

Use separate PostgreSQL databases:

```text
casdoor
  owned by Casdoor
  initialized by Casdoor setup/migrations

ivyvue
  owned by the control-plane application
  initialized by our migrations
```

Local bootstrap role:

```text
role: jaten
password: jaten
purpose: local bootstrap only
```

Runtime roles should be narrower:

```text
ivyvue_app
  runtime application role

ivyvue_migrator
  schema migration role

ivyvue_admin
  optional local/admin maintenance role
```

The `jaten` role is not the runtime application credential.

## Security Boundaries

Public/browser-reachable:

- control-plane app routes
- Casdoor public login/OIDC routes

Private/internal only:

- Casdoor admin/management API
- analysis server internal API
- control-plane admin command credentials
- database credentials

The analysis server should be reachable only from the control-plane server in
production. If it must bind localhost during development, keep that explicit.

The Casdoor admin API must never be exposed to browser JavaScript.

## Z3 And Future Compute Modes

For Phase 0, the analysis server keeps embedded Z3 via CGO.

Reasons:

- it is the working implementation today
- it avoids solving WASM/Z3 immediately
- it lets the Vue/control/auth architecture progress independently

Future analysis implementations should hide behind an engine boundary:

```text
AnalysisEngine
  CreateSession
  LoadModel
  CheckInduction
  RunAction
  FetchARG
  FetchConcept
  FetchMenus
  StreamEvents
  Cancel
```

Possible future engines:

- existing server process with embedded Z3
- server worker pool
- browser WASM trial mode
- Wanix browser runtime
- paid/isolated compute workers

The control-plane API should not care which engine executes a project/session.

## Local Development Topology

Suggested ports:

```text
127.0.0.1:18080  control-plane server
127.0.0.1:18081  ivyvue analysis server
127.0.0.1:18082  Casdoor public endpoint
127.0.0.1:5432   PostgreSQL
```

The exact ports can change, but the roles should remain clear.

Startup order:

```text
1. PostgreSQL
2. Casdoor
3. ivyvue analysis server
4. Go control-plane server
```

Local bootstrap should create two databases:

```text
casdoor
  used only by Casdoor

ivyvue
  used by the control-plane server
```

The local PostgreSQL `jaten`/`jaten` role may be used to bootstrap both
databases. It should not be the runtime credential for either production
service.

## BDD Testing Plan

Use behavior-driven tests as the first executable specification for auth,
billing, and tester onboarding. The tests should describe user-visible behavior
in Given/When/Then language, then bind those scenarios to Playwright browser
tests and Go API/store tests.

The BDD suite should cover these flows before implementation:

```text
Sign-up
Login
Account recovery
Billing information
Alpha tester creation from the dashboard
```

### Test Layers

Use three layers so the suite is both clear and fast:

```text
Feature specs
  Gherkin-style .feature or markdown scenarios; source of truth for behavior.

Browser specs
  Playwright tests that exercise the control-plane UI and Casdoor UI.

Go integration specs
  database/API tests for edge cases that are expensive or brittle in a browser.
```

Suggested layout:

```text
goivy/control/features/
  signup.feature
  login.feature
  account_recovery.feature
  billing.feature
  alpha_dashboard.feature

goivy/control/pw_test/
  signup.spec.ts
  login.spec.ts
  account_recovery.spec.ts
  billing.spec.ts
  alpha_dashboard.spec.ts

goivy/control/auth_test.go
goivy/control/billing_test.go
goivy/control/dashboard_test.go
goivy/control/testsupport/
  email_sink.go
  fake_billing_provider.go
  casdoor_fixture.go
```

Do not require Cucumber on day one unless it proves helpful. It is enough for
test names and comments to mirror Given/When/Then steps exactly. The important
part is test-first behavioral coverage, not a specific BDD runner.

### Shared Test Infrastructure

BDD tests need deterministic local fixtures:

- disposable PostgreSQL test database or schema for `ivyvue`
- disposable or resettable Casdoor test organization/application
- fake SMTP/email sink that captures verification and recovery emails
- fake billing provider that records customer/payment-method calls without
  contacting a real payment service
- seeded admin user who can open the control-plane dashboard
- seeded project template for alpha testers
- browser test helper that can follow login redirects across control-plane and
  Casdoor origins

The fake billing provider should be an implementation of a small interface:

```text
BillingProvider
  CreateCustomer(account, billingEmail)
  CreateSetupIntent(account)
  AttachPaymentMethod(account, paymentMethodToken)
  MarkDefaultPaymentMethod(account, paymentMethodID)
```

Production can later use Stripe or another provider. BDD tests should not care
which provider is selected.

### Sign-Up Feature

Goal: a new person can register with an email address, verify that email, and
land in a usable starter account/project.

Primary scenario:

```gherkin
Feature: Sign-up

  Scenario: New user signs up and verifies email
    Given no product user exists for "alice@example.test"
    And the browser is on the control-plane home page
    When Alice chooses to sign up
    And Alice enters "alice@example.test" and a valid password in Casdoor
    Then Casdoor sends an email verification message to "alice@example.test"
    When Alice opens the verification link from the test email sink
    And Alice returns to the control-plane application
    Then Alice is logged in
    And a control.users row exists for Alice's Casdoor issuer and subject
    And Alice has a billing account
    And Alice has a starter project
    And Alice can open the project workspace
```

Negative scenarios:

```gherkin
Scenario: Duplicate sign-up does not create duplicate product users
Scenario: Unverified email cannot access project workspace if verification is required
Scenario: Invalid email is rejected before product state is created
Scenario: Weak password is rejected by Casdoor policy
Scenario: Sign-up cancellation returns to the public control-plane page
```

Assertions:

- `control.users` stores `idp_issuer`, `idp_subject`, email, display name, and
  email verification state.
- duplicate retries are idempotent.
- no billing account/project is created for failed registration unless the
  product intentionally supports pending accounts.

### Login Feature

Goal: a returning user can authenticate through Casdoor and receive a
control-plane app session.

Primary scenario:

```gherkin
Feature: Login

  Scenario: Existing user logs in and sees authorized projects
    Given Alice has a verified Casdoor user
    And Alice has access to project "dev/client-server"
    When Alice logs in through Casdoor
    Then the control-plane creates an HttpOnly app session cookie
    And /auth/me returns Alice's user profile
    And /auth/me returns Alice's billing accounts, teams, projects, and roles
    And Alice can open "dev/client-server"
```

Negative scenarios:

```gherkin
Scenario: Wrong password does not create an app session
Scenario: Disabled Casdoor user cannot create an app session
Scenario: Product-disabled user cannot access projects after Casdoor login succeeds
Scenario: User without project grants sees an empty project picker
Scenario: Logout clears the app session and returns to the unauthenticated state
Scenario: Browser refresh preserves the app session until idle or absolute expiry
```

Assertions:

- browser JavaScript cannot read `ivy_webvue_session`.
- no Casdoor admin credential reaches the browser.
- no raw long-lived Casdoor token is stored in localStorage/sessionStorage.
- project APIs reject requests after logout.

### Account-Recovery Feature

Goal: a user who lost their password can recover access through email without
leaking whether arbitrary emails are registered.

Primary scenario:

```gherkin
Feature: Account recovery

  Scenario: User resets password from a recovery email
    Given Alice has a verified Casdoor user
    When Alice requests password recovery for "alice@example.test"
    Then the page shows a neutral recovery response
    And the test email sink receives a recovery email for Alice
    When Alice opens the recovery link
    And Alice sets a new valid password
    Then Alice can log in with the new password
    And Alice cannot log in with the old password
```

Negative scenarios:

```gherkin
Scenario: Unknown email receives the same neutral response
Scenario: Expired recovery link is rejected
Scenario: Used recovery link cannot be reused
Scenario: Weak replacement password is rejected
Scenario: Recovery for disabled user does not reactivate product access
```

Assertions:

- recovery responses do not reveal whether an email exists.
- recovery links are consumed once.
- successful recovery invalidates existing app sessions if that is the chosen
  policy.

### Billing Information Feature

Goal: a billing owner/admin can add billing information for an account without
granting project access accidentally.

Phase 0 should test against a fake billing provider. The UI can be real, but
payment-provider calls are captured by the fake adapter.

Primary scenario:

```gherkin
Feature: Billing information

  Scenario: Account owner adds billing information
    Given Alice is an owner of billing account "acme"
    And "acme" has no default payment method
    When Alice opens account billing settings
    And Alice enters valid billing contact information
    And Alice enters a valid test payment method
    Then the billing provider receives a create-customer request for "acme"
    And the billing provider receives an attach-payment-method request
    And "acme" has billing_status "active" or "payment_method_on_file"
    And no project grants are changed
```

Negative scenarios:

```gherkin
Scenario: Account member cannot open billing settings
Scenario: Billing admin can update billing but receives no project access
Scenario: Payment method failure leaves the account in actionable billing state
Scenario: Refreshing during billing setup does not double-create customers
Scenario: Billing email changes are audited
```

Assertions:

- only `owner`, `admin`, or `billing_admin` can manage billing.
- `billing_admin` does not imply project access.
- billing writes are idempotent under retry.
- no raw card data is stored in `ivyvue`.

### Alpha Dashboard Feature

Goal: an admin can create or invite alpha testers from our dashboard and grant
them project access without asking them to create Gmail/GitHub accounts.

Primary scenario:

```gherkin
Feature: Alpha tester dashboard

  Scenario: Admin invites an alpha tester and grants project access
    Given Admin is logged into the control-plane dashboard
    And project "alpha/client-server" exists
    When Admin creates alpha tester "tester1@example.test"
    And Admin grants tester1 read/write access to "alpha/client-server"
    Then Casdoor has a user or invitation for "tester1@example.test"
    And control.users has a pending or active mapped user record
    And the tester is an account user in the alpha billing account
    And the tester has write access to "alpha/client-server"
    And an invitation email is sent to "tester1@example.test"
```

Tester acceptance scenario:

```gherkin
Scenario: Invited alpha tester accepts invite and opens project
  Given Admin invited "tester1@example.test"
  When Tester opens the invitation email
  And Tester completes Casdoor account setup
  And Tester returns to the control-plane application
  Then Tester sees "alpha/client-server"
  And Tester can open the Ivy workspace
  And Tester cannot open projects they were not granted
```

Negative scenarios:

```gherkin
Scenario: Non-admin cannot create alpha testers
Scenario: Creating the same alpha tester twice is idempotent
Scenario: Revoked tester loses project access immediately
Scenario: Dashboard-created tester can be assigned to a team
Scenario: Dashboard-created tester can be marked alpha or beta for segmentation
```

Assertions:

- dashboard actions call Casdoor management APIs only from the server.
- dashboard actions update `ivyvue` product authorization in the same logical
  workflow.
- partial failures are visible and retryable.
- alpha/beta flags live in `ivyvue`, not only in Casdoor.

### BDD Implementation Order

Write tests in this order:

1. Create feature files for the five flows above.
2. Add test support: disposable DB, fake email sink, fake billing provider, and
   Casdoor fixture helpers.
3. Write failing Playwright scenario for login with a seeded Casdoor user.
4. Implement the minimum control-plane OIDC callback/session code.
5. Write failing `/auth/me` API tests for user/account/team/project visibility.
6. Implement user mapping and project visibility.
7. Write failing sign-up and email verification scenarios.
8. Implement sign-up integration policy and product bootstrap behavior.
9. Write failing account recovery scenarios using the email sink.
10. Configure/bridge Casdoor recovery behavior and assert app-session results.
11. Write failing billing information scenarios against fake billing provider.
12. Implement billing account settings and provider adapter.
13. Write failing alpha dashboard scenarios.
14. Implement dashboard admin workflows and Casdoor provisioning calls.
15. Add negative/authorization scenarios for all five features.
16. Add CI grouping so fast API/store tests run before slower browser flows.

### BDD Definition Of Done

The BDD plan is complete when:

- every flow has at least one red browser scenario before implementation
- every flow has negative authorization/error scenarios
- tests can run without external Gmail/GitHub accounts
- tests can run without real payment-provider network calls
- emails are captured and inspected by the test suite
- Casdoor admin/API credentials never enter browser tests except through server
  behavior
- project access is asserted through both UI behavior and server-side API
  rejection
- BDD test names remain readable enough for product discussions

## Implementation Order

1. Document process topology and local ports.
2. Create PostgreSQL bootstrap scripts for `casdoor` and `ivyvue` roles and
   schemas.
3. Add Casdoor local setup and configuration documentation.
4. Write the BDD feature files for sign-up, login, account recovery, billing,
   and alpha dashboard flows.
5. Add failing BDD-backed tests and local test fixtures before production flow
   implementation.
6. Create the Go control-plane command.
7. Serve a minimal Vue app from the control-plane server.
8. Add OIDC login/callback/logout against Casdoor.
9. Add control-plane app session cookie.
10. Add `control.users` mapping from Casdoor/OIDC `issuer + subject`.
11. Add billing account, account user, team, team membership, project, and grant
   migrations.
12. Add project authorization service and tests.
13. Add PostgreSQL RLS helper and tests.
14. Split or wrap the existing Ivy/Z3 web backend as an internal analysis
    server.
15. Add the control-plane analysis client.
16. Add project-scoped public Ivy APIs in the control-plane server.
17. Connect Vue stores/components to control-plane APIs.
18. Add admin command for tester invite/account/team/project grant workflows
    that calls Casdoor APIs and updates `ivyvue`.
19. Add browser smoke tests for login, project picker, and opening an analysis
    session.
20. Add tests that prove the browser cannot access analysis without a project
    grant.

## Phase 0 Definition Of Done

Phase 0 is complete when:

- browser starts at the control-plane server
- BDD specs exist for sign-up, login, account recovery, billing information,
  and alpha dashboard tester creation
- unauthenticated users are sent through Casdoor login
- successful login creates an app session cookie
- Casdoor/OIDC user maps to `control.users`
- the seeded/dev user can see a billing account, team, and project
- project access is enforced by the control-plane server
- project-owned data uses PostgreSQL RLS
- control-plane can create an analysis session in the ivyvue analysis server
- Vue can show the workspace shell for an authorized project
- no browser route talks directly to the analysis server
- admin tooling can create/invite alpha or beta testers and grant project
  access

## Open Decisions

- Whether Casdoor runs on a separate host/domain or under the same domain
  behind path-based routing.
- Whether the control-plane server or a reverse proxy terminates TLS.
- Exact internal credential format for control-plane to analysis server.
- Whether analysis session state is entirely in memory at first or persisted in
  `control.ivy_workspace_sessions`.
- How quickly to add queueing, timeouts, and cancellation around Z3 work.
- Whether trial users start with server-backed Z3 or an eventual browser/WASM
  engine.

## Casdoor Spike Acceptance Criteria

The Casdoor spike is successful when:

- Casdoor runs locally against PostgreSQL.
- The control-plane server can redirect to Casdoor login.
- Casdoor redirects back to `/auth/callback`.
- The control-plane server validates the OIDC response.
- The control-plane server creates `ivy_webvue_session`.
- `control.users` stores `idp_issuer`, `idp_subject`, email, display name, and
  email verification state.
- A tester can be created or invited without requiring Gmail/GitHub accounts.
- An admin workflow can create a billing account, team, project, and grant for
  that tester.
- The Vue app can call `/auth/me` and see the authorized project.
- The browser never receives Casdoor admin credentials or raw long-lived
  Casdoor tokens.

If any of these fail because Casdoor lacks a needed capability or is too hard
to operate, revisit the fallback list:

```text
1. Ory Kratos + Ory Hydra
2. Keycloak
3. minimal local email/password auth
```
