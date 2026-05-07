Feature: Login

  Scenario: Existing user logs in and sees authorized projects
    Given Alice has a verified product user
    And Alice has access to project "dev/client-server"
    When Alice requests a login link for "alice@example.test"
    And Alice opens the magic link within 10 minutes
    Then the control-plane creates an HttpOnly app session cookie
    And /auth/me returns Alice's user profile
    And /auth/me returns Alice's billing accounts, teams, projects, and roles
    And Alice can open "dev/client-server"

  Scenario: Reused magic link does not create another app session
    Given Alice requested a login link
    And Alice already used the login link
    When Alice opens the same link again
    Then the control-plane rejects the token
    And no new app session cookie is created

  Scenario: Expired magic link does not create an app session
    Given Alice requested a login link more than 10 minutes ago
    When Alice opens the magic link
    Then the control-plane rejects the token
    And /auth/me returns unauthenticated

  Scenario: Product-disabled user cannot access projects after email verification succeeds
    Given Alice has a verified product user
    And Alice is disabled in control.users
    When Alice opens a fresh magic link
    Then the control-plane rejects the session
    And Alice cannot access project APIs

  Scenario: User without project grants sees an empty project picker
    Given Alice has a verified product user
    And Alice has no project grants
    When Alice logs in by email magic link
    Then Alice sees an empty project picker
    And project-scoped Ivy APIs reject Alice

  Scenario: Logout clears the app session and returns to the unauthenticated state
    Given Alice is logged in
    When Alice logs out
    Then the control-plane clears the app session cookie
    And /auth/me returns unauthenticated

  Scenario: Browser return refreshes the app session for 72 hours
    Given Alice is logged in
    When Alice returns to the website within 72 hours
    Then Alice remains logged in
    And the app session expiry is refreshed for another 72 hours
    When Alice returns after the app session expires
    Then /auth/me returns unauthenticated
