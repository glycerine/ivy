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

  Scenario: Unknown email receives the same neutral response
    Given no user exists for "unknown@example.test"
    When someone requests password recovery for "unknown@example.test"
    Then the page shows the same neutral recovery response
    And the response does not reveal whether the email exists

  Scenario: Expired recovery link is rejected
    Given Alice has an expired recovery link
    When Alice opens the recovery link
    Then Casdoor rejects the recovery link
    And Alice cannot set a new password from that link

  Scenario: Used recovery link cannot be reused
    Given Alice used a recovery link successfully
    When Alice opens the same recovery link again
    Then Casdoor rejects the recovery link

  Scenario: Weak replacement password is rejected
    Given Alice opens a valid recovery link
    When Alice enters a weak replacement password
    Then Casdoor rejects the password
    And the old password remains unchanged

  Scenario: Recovery for disabled user does not reactivate product access
    Given Alice is disabled in control.users
    When Alice completes password recovery in Casdoor
    Then Alice still cannot access product projects
