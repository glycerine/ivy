# Ivy Control Plane

This package is the start of the browser-facing control-plane server described
in `../../architecture_with_auth.md`.

Current responsibilities in this first slice:

- serve a minimal browser entrypoint
- expose `/auth/me`
- start the Casdoor OIDC login redirect at `/auth/login`
- exchange and validate OIDC callback ID tokens
- create the HttpOnly `ivy_webvue_session` app cookie
- map OIDC issuer+subject to local product users
- return the authenticated user's starter account, team, project, and role
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

The browser smoke test uses `IVY_CONTROL_TEST_IDP=1`, which enables an
in-process deterministic OIDC issuer. Real local Casdoor/PostgreSQL setup notes
live under `admin/`.

Full sign-up, account recovery, billing, and alpha-dashboard scenarios are
currently captured as BDD specs and Playwright `fixme` placeholders until those
product flows are implemented.
