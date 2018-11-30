@replies
Feature: Thread Reply support

  Scenario: User replies to message with threading disabled

  Scenario: User replies to uncached message with unthreading

  Scenario: User replies to unthreaded message

  Scenario: User replies to own message
    Given I have a simple gateway between #foo and #bar
    And I post a message in #foo
    And I post a reply in #foo
    Then a reply should appear in #bar

  Scenario: User replies to relayed message
    Given I have a simple gateway between #foo and #bar
    And I post a message in #foo
    And I post a reply in #bar
    Then a reply should appear in #foo

  Scenario: User edits a reply message
    Given I have a simple gateway between #foo and #bar
    And I post a message in #foo
    And I post a reply in #bar
    And I edit the reply
    Then the edited reply should appear in #foo

  Scenario: User deletes a reply message
    Given I have a simple gateway between #foo and #bar
    And I post a message in #foo
    And I post a reply in #bar
    And I delete the reply
    Then the deleted reply should not appear in #foo

  Scenario: User deletes a parent message
    Given I have a simple gateway between #foo and #bar
    And I post a message in #foo
    And I post a reply in #bar
    And I delete the parent message in #foo
    Then the parent message in #bar should be "This message was deleted."

  Scenario: User deletes a parent message, then replies to original parent
    Given I have a simple gateway between #foo and #bar
    And I post a message in #foo
    And I post a reply in #bar
    And I delete the parent message in #foo
    And I post another reply in #foo
    Then the final reply should appear in #bar

  Scenario: User deletes a parent message, then replies to relayed parent
    Given I have a simple gateway between #foo and #bar
    And I post a message in #foo
    And I post a reply in #bar
    And I delete the parent message in #foo
    And I post another reply in #bar
    Then the final reply should appear in #foo
