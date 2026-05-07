Feature: Alpha tester dashboard

  Scenario: Admin invites an alpha tester and grants project access
    Given Admin is logged into the control-plane dashboard
    And project "alpha/client-server" exists
    When Admin creates alpha tester "tester1@example.test"
    And Admin grants tester1 read/write access to "alpha/client-server"
    Then Casdoor has a user or invitation for "tester1@example.test"
    And control.users has a pending or active mapped user record
    And the tester is an account user in the alpha billing account
    And the tester has write access to "alpha/client-server"
    And an invitation email is sent to "tester1@example.test"

  Scenario: Invited alpha tester accepts invite and opens project
    Given Admin invited "tester1@example.test"
    When Tester opens the invitation email
    And Tester completes Casdoor account setup
    And Tester returns to the control-plane application
    Then Tester sees "alpha/client-server"
    And Tester can open the Ivy workspace
    And Tester cannot open projects they were not granted

  Scenario: Non-admin cannot create alpha testers
    Given Alice is not a dashboard admin
    When Alice tries to create an alpha tester
    Then the control-plane rejects the request
    And Casdoor is not called

  Scenario: Creating the same alpha tester twice is idempotent
    Given Admin already invited "tester1@example.test"
    When Admin creates alpha tester "tester1@example.test" again
    Then there is one Casdoor user or invitation
    And there is one control.users row
    And the requested project grant exists

  Scenario: Revoked tester loses project access immediately
    Given tester1 has write access to "alpha/client-server"
    When Admin revokes tester1's access
    Then tester1 cannot open "alpha/client-server"

  Scenario: Dashboard-created tester can be assigned to a team
    Given Admin created alpha tester "tester1@example.test"
    When Admin adds tester1 to team "alpha/core"
    Then tester1 inherits the team's project grants

  Scenario: Dashboard-created tester can be marked alpha or beta for segmentation
    Given Admin created tester "tester1@example.test"
    When Admin marks tester1 as "alpha"
    Then ivyvue stores tester1's alpha segmentation flag
