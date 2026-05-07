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

  Scenario: Duplicate sign-up does not create duplicate product users
    Given Alice already has a Casdoor identity and control.users row
    When Alice repeats the sign-up flow with the same email
    Then exactly one control.users row exists for Alice
    And exactly one starter billing account exists for Alice
    And the flow is safe to retry

  Scenario: Unverified email cannot access project workspace if verification is required
    Given Alice registered but did not verify email
    When Alice returns to the control-plane application
    Then Alice sees an email verification required state
    And Alice cannot open a project workspace

  Scenario: Invalid email is rejected before product state is created
    Given no product user exists for "not-an-email"
    When Alice attempts to sign up with "not-an-email"
    Then Casdoor rejects the email
    And no control.users row is created

  Scenario: Weak password is rejected by Casdoor policy
    Given no product user exists for "weak@example.test"
    When Alice attempts to sign up with a weak password
    Then Casdoor rejects the password
    And no starter project is created

  Scenario: Sign-up cancellation returns to the public control-plane page
    Given the browser is in the Casdoor sign-up flow
    When Alice cancels sign-up
    Then Alice returns to the unauthenticated control-plane state
