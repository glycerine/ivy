Feature: Login

  Scenario: Existing user logs in and sees authorized projects
    Given Alice has a verified Casdoor user
    And Alice has access to project "dev/client-server"
    When Alice logs in through Casdoor
    Then the control-plane creates an HttpOnly app session cookie
    And /auth/me returns Alice's user profile
    And /auth/me returns Alice's billing accounts, teams, projects, and roles
    And Alice can open "dev/client-server"

  Scenario: Wrong password does not create an app session
    Given Alice has a verified Casdoor user
    When Alice attempts to log in with the wrong password
    Then Casdoor rejects the login
    And the control-plane does not create an app session cookie

  Scenario: Disabled Casdoor user cannot create an app session
    Given Alice is disabled in Casdoor
    When Alice attempts to log in
    Then login fails
    And /auth/me returns unauthenticated

  Scenario: Product-disabled user cannot access projects after Casdoor login succeeds
    Given Alice can log in through Casdoor
    And Alice is disabled in control.users
    When Alice returns to the control-plane callback
    Then the control-plane rejects the session
    And Alice cannot access project APIs

  Scenario: User without project grants sees an empty project picker
    Given Alice has a verified Casdoor user
    And Alice has no project grants
    When Alice logs in through Casdoor
    Then Alice sees an empty project picker
    And project-scoped Ivy APIs reject Alice

  Scenario: Logout clears the app session and returns to the unauthenticated state
    Given Alice is logged in
    When Alice logs out
    Then the control-plane clears the app session cookie
    And /auth/me returns unauthenticated

  Scenario: Browser refresh preserves the app session until idle or absolute expiry
    Given Alice is logged in
    When Alice refreshes the browser
    Then Alice remains logged in
    When the app session expires
    Then /auth/me returns unauthenticated
