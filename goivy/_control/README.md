# Ivy Control Plane

This package is the start of the browser-facing control-plane server described
in `../../architecture_with_auth.md`.

Current responsibilities in this slice:

- materialize the embedded Vue browser app into `./.runweb` at startup
- serve the browser entrypoint and static assets from that disk directory
- expose `/auth/me`
- request email magic links without revealing whether an email exists
- send sign-in links through an `EmailSender` interface
- record local/test sign-in links in the `email_deliveries` table so no paid
  email is sent
- optionally send real email through Mailgun when an API key is explicitly set
- expose an `/admin` dashboard for pending unverified email links
- consume 10-minute, single-use email login tokens
- create and refresh the HttpOnly `ivy_webui_session` app cookie for 400 days
- record authenticated return visits in hourly `visiting_hours` rows
- return the authenticated user's starter account, team, project, and role
- keep OIDC client code available for future Google/GitHub sign-in attachment
- define the product model for users, billing accounts, teams, projects, and
  grants
- hold BDD feature specs for auth, billing, and alpha tester flows

Run targeted Go tests:

```sh
cd goivy
go test ./control ./control/testsupport ./cmd/ivy-control ./cmd/ivy-control-admin
```

Run the browser BDD smoke tests:

```sh
cd goivy/control
npm ci
npm run test:browser
```

The Playwright server binds localhost and may require sandbox approval in Codex.

The browser smoke test uses `IVY_CONTROL_TEST_EMAIL_OUTBOX=1`, which enables a
test-only endpoint for the in-memory email outbox. Real Mailgun sending is not
used by tests.

Run the browser-facing server locally:

```sh
cd goivy
make build-vue
IVY_CONTROL_DATABASE_DSN='postgres://ivyvue_app:ivyvue_app_dev@127.0.0.1:5432/ivyvue?sslmode=disable' \
IVY_CONTROL_TEST_EMAIL_OUTBOX=1 \
go run ./cmd/ivy-control
```

On startup, `ivy-control` copies the embedded `webui` assets into `./.runweb`
relative to the directory where the binary is run, then serves those files. The
binary therefore does not need to be launched from the repository. While the
server is running, edits to files in `.runweb` are served directly. Use
`-static-dir` or `IVY_CONTROL_STATIC_DIR` only when you explicitly want to serve
some other asset directory.

Full account recovery, billing, and alpha-dashboard scenarios are currently
captured as BDD specs and Playwright `fixme` placeholders until those product
flows are implemented.
