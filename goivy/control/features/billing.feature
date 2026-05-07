Feature: Billing information

  Scenario: Account owner adds billing information
    Given Alice is an owner of billing account "acme"
    And "acme" has no default payment method
    When Alice opens account billing settings
    And Alice enters valid billing contact information
    And Alice enters a valid test payment method
    Then the billing provider receives a create-customer request for "acme"
    And the billing provider receives an attach-payment-method request
    And "acme" has billing_status "active" or "payment_method_on_file"
    And no project grants are changed

  Scenario: Account member cannot open billing settings
    Given Alice is a member of billing account "acme"
    When Alice opens account billing settings
    Then the control-plane rejects the request

  Scenario: Billing admin can update billing but receives no project access
    Given Alice is a billing_admin of billing account "acme"
    When Alice adds valid billing information
    Then the billing information is saved
    And Alice receives no project access from the billing role alone

  Scenario: Payment method failure leaves the account in actionable billing state
    Given Alice is an owner of billing account "acme"
    And the fake billing provider is configured to fail payment attachment
    When Alice enters a test payment method
    Then "acme" remains in an actionable billing state
    And Alice sees a retryable error

  Scenario: Refreshing during billing setup does not double-create customers
    Given Alice is adding billing information for "acme"
    When Alice refreshes during billing setup
    Then the billing provider has at most one customer for "acme"

  Scenario: Billing email changes are audited
    Given Alice is an owner of billing account "acme"
    When Alice changes the billing email
    Then an audit record captures the old and new billing email
