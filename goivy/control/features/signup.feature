Feature: Sign-up

  Scenario: New user signs up and verifies email
    Given no product user exists for "alice@example.test"
    And the browser is on the control-plane home page
    When Alice enters "alice@example.test" to request a sign-up link
    Then the page shows a neutral email-sent response
    And the fake email sink receives a magic link for "alice@example.test"
    When Alice opens the magic link within 10 minutes
    Then Alice is logged in
    And a control.users row exists for Alice's verified email identity
    And Alice has a billing account
    And Alice has a starter project
    And Alice can open the project workspace

  Scenario: Duplicate sign-up does not create duplicate product users
    Given Alice already has a verified control.users row
    When Alice requests another sign-up link with the same email
    And Alice opens the new magic link
    Then exactly one control.users row exists for Alice
    And exactly one starter billing account exists for Alice
    And the flow is safe to retry

  Scenario: Unused magic link is required before project access
    Given Alice requested a sign-up link but did not open it
    When Alice returns to the control-plane application
    Then Alice sees an unauthenticated state
    And Alice cannot open a project workspace

  Scenario: Invalid email is rejected before product state is created
    Given no product user exists for "not-an-email"
    When Alice attempts to sign up with "not-an-email"
    Then no control.users row is created
    And no magic link is sent

  Scenario: Sign-up link cancellation returns to the public control-plane page
    Given Alice requested a sign-up link
    When Alice ignores the email
    Then Alice remains in the unauthenticated control-plane state
