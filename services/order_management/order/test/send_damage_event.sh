#!/bin/bash
# Test script to send order damage events to RabbitMQ using curl and RabbitMQ HTTP API

# Default values
SEVERITY=${1:-minor}
ORDER_ID=${2:-test_order_$(date +%s)}
RABBITMQ_HOST=${RABBITMQ_HOST:-localhost}
RABBITMQ_PORT=${RABBITMQ_PORT:-15672}
RABBITMQ_USER=${RABBITMQ_USER:-guest}
RABBITMQ_PASS=${RABBITMQ_PASS:-guest}

echo "🚀 Sending $SEVERITY damage event for order: $ORDER_ID"

# Create the damage event payload
DAMAGE_PAYLOAD=$(cat <<EOF
{
  "eventId": "evt_$(date +%s)",
  "type": "order.damage",
  "source": "test-script",
  "occurredAt": "$(date -u +%Y-%m-%dT%H:%M:%S.%3NZ)",
  "orderId": "$ORDER_ID",
  "severity": "$SEVERITY",
  "description": "Test damage event: severity=$SEVERITY, temp=12.3C, humidity=70%",
  "details": {
    "temperature": 12.3,
    "humidity": 70,
    "status": "active",
    "mqttTopic": "events/sensor"
  }
}
EOF
)

# Create the MQTT message wrapper
MQTT_MESSAGE=$(cat <<EOF
{
  "mqtt_topic": "events/order-damage",
  "payload": "$(echo "$DAMAGE_PAYLOAD" | sed 's/"/\\"/g' | tr -d '\n')",
  "timestamp": $(date +%s.%3N)
}
EOF
)

echo "📄 Message payload:"
echo "$MQTT_MESSAGE" | jq .

# Send message to RabbitMQ using HTTP API
RESPONSE=$(curl -s -u "$RABBITMQ_USER:$RABBITMQ_PASS" \
  -H "Content-Type: application/json" \
  -X POST \
  "http://$RABBITMQ_HOST:$RABBITMQ_PORT/api/exchanges/%2F/events/publish" \
  -d "{
    \"properties\": {
      \"delivery_mode\": 2,
      \"content_type\": \"application/json\"
    },
    \"routing_key\": \"order.damage\",
    \"payload\": \"$(echo "$MQTT_MESSAGE" | base64 -w 0)\",
    \"payload_encoding\": \"base64\"
  }")

if echo "$RESPONSE" | grep -q '"routed":true'; then
  echo "✅ Message sent successfully!"
else
  echo "❌ Failed to send message. Response: $RESPONSE"
  exit 1
fi

echo "🎉 Test completed!"
echo ""
echo "💡 Check the order service logs with:"
echo "   docker logs order-management -f"
echo ""
echo "💡 Check RabbitMQ management UI at:"
echo "   http://localhost:15672 (guest/guest)"