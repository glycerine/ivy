# Ivy Control Plane

This package is the start of the browser-facing control-plane server described
in `../../architecture_with_auth.md`.

Current responsibilities in this slice:

- serve a minimal browser entrypoint
- expose `/auth/me`
- request email magic links without revealing whether an email exists
- send sign-in links through an `EmailSender` interface
- use an in-memory email sender in tests so no paid email is sent
- optionally send real email through Mailgun when an API key is explicitly set
- consume 10-minute, single-use email login tokens
- create and refresh the HttpOnly `ivy_webvue_session` app cookie for 72 hours
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

Full account recovery, billing, and alpha-dashboard scenarios are currently
captured as BDD specs and Playwright `fixme` placeholders until those product
flows are implemented.
