# Order Damage Event Handling

This document describes the implementation of order damage event handling in the order management service.

## Overview

The order management service has been updated to handle order damage events received from MQTT topics. These events are generated when sensors detect potential damage to orders during transit or storage.

## Event Structure

### MQTT Message Format
```json
{
  "mqtt_topic": "events/order-damage",
  "payload": "{...}",
  "timestamp": 1759555253.2585864
}
```

### Order Damage Event Payload
```json
{
  "eventId": "evt_1759555253",
  "type": "order.damage",
  "source": "mqtt-order-event-client",
  "occurredAt": "2025-10-04T05:20:53.182048461Z",
  "orderId": "evt_1759555253",
  "severity": "minor",
  "description": "Potential damage detected: temp=9.23C, humidity=58.00%",
  "details": {
    "temperature": 9.234309268737501,
    "humidity": 58,
    "status": "active",
    "mqttTopic": "events/sensor"
  }
}
```

## Implementation Details

### Domain Models

1. **OrderDamageEvent**: Represents the damage event with all relevant information
2. **OrderDamageDetails**: Contains sensor data that triggered the damage detection
3. **MQTTOrderEvent**: Wrapper for MQTT messages containing nested JSON payloads

### Consumer Adapter Updates

The `OrderConsumerAdapter` has been enhanced to:
- Detect MQTT order damage events by topic (`events/order-damage`)
- Parse nested JSON payloads
- Route events to appropriate handlers based on event type

### Service Layer Updates

The `OrderService` now implements `HandleOrderDamageEvent` which:
- Logs damage event details
- Updates order status based on damage severity:
  - **Minor**: Sets status to `damage_detected_minor`
  - **Major**: Sets status to `damage_detected_major`
  - **Critical**: Cancels the order with status `cancelled_damage`

## Damage Severity Levels

| Severity | Action | Order Status |
|----------|--------|--------------|
| minor | Monitor and log | `damage_detected_minor` |
| major | Immediate attention required | `damage_detected_major` |
| critical | Automatic cancellation | `cancelled_damage` |

## Usage Example

See `examples/order_damage_example.go` for a complete example of how damage events are processed.

## Future Enhancements

The current implementation provides a foundation for:
- Automated notifications to warehouse staff
- Integration with insurance claim systems
- Damage report generation
- Inventory status updates
- Customer notifications

## Testing

To test the damage event handling:

1. Send an MQTT message to the `events/order-damage` topic
2. Verify the message is consumed by the order service
3. Check that the order status is updated appropriately
4. Review logs for damage event processing details