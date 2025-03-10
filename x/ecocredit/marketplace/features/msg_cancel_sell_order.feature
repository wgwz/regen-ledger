Feature: MsgCancelSellOrder

  Scenario: a valid message
    Given the message
    """
    {
      "seller": "regen1elq7ys34gpkj3jyvqee0h6yk4h9wsfxmgqelsw",
      "sell_order_id": 1
    }
    """
    When the message is validated
    Then expect no error

  Scenario: an error is returned if seller is empty
    Given the message
    """
    {}
    """
    When the message is validated
    Then expect the error "seller cannot be empty: invalid request"

  Scenario: an error is returned if seller is not a valid bech32 address
    Given the message
    """
    {
      "seller": "foo"
    }
    """
    When the message is validated
    Then expect the error "seller is not a valid address: decoding bech32 failed: invalid bech32 string length 3: invalid address"

  Scenario: an error is returned if sell order id is empty
    Given the message
    """
    {
      "seller": "regen1elq7ys34gpkj3jyvqee0h6yk4h9wsfxmgqelsw"
    }
    """
    When the message is validated
    Then expect the error "sell order id cannot be empty: invalid request"
