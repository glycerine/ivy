Feature: Account recovery

  Scenario: User recovers access with a fresh email magic link
    Given Alice has a verified product user
    When Alice requests account recovery for "alice@example.test"
    Then the page shows a neutral recovery response
    And the fake email sink receives a magic link for Alice
    When Alice opens the magic link within 10 minutes
    Then Alice is logged in
    And Alice can access her authorized projects

  Scenario: Unknown email receives the same neutral response
    Given no user exists for "unknown@example.test"
    When someone requests account recovery for "unknown@example.test"
    Then the page shows the same neutral recovery response
    And the response does not reveal whether the email exists

  Scenario: Expired recovery link is rejected
    Given Alice has an expired recovery link
    When Alice opens the recovery link
    Then the control-plane rejects the token
    And Alice remains unauthenticated

  Scenario: Used recovery link cannot be reused
    Given Alice used a recovery link successfully
    When Alice opens the same recovery link again
    Then the control-plane rejects the token

  Scenario: Recovery for disabled user does not reactivate product access
    Given Alice is disabled in users
    When Alice opens a fresh recovery link
    Then Alice still cannot access product projects
