# Ivy Control Plane

This package is the start of the browser-facing control-plane server described
in `../../architecture_with_auth.md`.

Current responsibilities in this first slice:

- serve a minimal browser entrypoint
- expose `/auth/me`
- start the Casdoor OIDC login redirect at `/auth/login`
- reject mismatched OIDC callback state
- define the product model for users, billing accounts, teams, projects, and
  grants
- hold BDD feature specs for auth, billing, and alpha tester flows

Run targeted Go tests:

```sh
cd goivy
go test ./control ./control/testsupport ./cmd/ivy-control ./cmd/ivy-control-admin
```

Run the initial browser BDD smoke tests:

```sh
cd goivy/control
npm ci
npm run test:browser
```

The Playwright server binds localhost and may require sandbox approval in Codex.

Casdoor is not required for the first login redirect test. Full sign-up,
account recovery, billing, and alpha-dashboard scenarios are currently captured
as BDD specs and Playwright `fixme` placeholders until the Casdoor and dashboard
integration is implemented.
