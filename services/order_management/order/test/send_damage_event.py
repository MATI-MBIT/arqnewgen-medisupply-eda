#!/usr/bin/env python3
"""
Test script to send order damage events to RabbitMQ
Requires: pip install pika
"""

import pika
import json
import sys
from datetime import datetime, timezone

def send_damage_event(severity="minor", order_id=None):
    """Send an order damage event to RabbitMQ"""
    
    # Default order ID if not provided
    if not order_id:
        order_id = f"test_order_{int(datetime.now().timestamp())}"
    
    # Create the MQTT-style message structure
    damage_event_payload = {
        "eventId": f"evt_{int(datetime.now().timestamp())}",
        "type": "order.damage",
        "source": "test-script",
        "occurredAt": datetime.now(timezone.utc).isoformat(),
        "orderId": order_id,
        "severity": severity,
        "description": f"Test damage event: severity={severity}, temp=10.5C, humidity=65%",
        "details": {
            "temperature": 10.5,
            "humidity": 65,
            "status": "active",
            "mqttTopic": "events/sensor"
        }
    }
    
    # Wrap in MQTT message format
    mqtt_message = {
        "mqtt_topic": "events/order-damage",
        "payload": json.dumps(damage_event_payload),
        "timestamp": datetime.now().timestamp()
    }
    
    try:
        # Connect to RabbitMQ
        connection = pika.BlockingConnection(
            pika.ConnectionParameters(host='localhost', port=5672, 
                                    credentials=pika.PlainCredentials('guest', 'guest'))
        )
        channel = connection.channel()
        
        # Publish the message
        channel.basic_publish(
            exchange='events',
            routing_key='order.damage',
            body=json.dumps(mqtt_message),
            properties=pika.BasicProperties(
                delivery_mode=2,  # Make message persistent
                content_type='application/json'
            )
        )
        
        print(f"✅ Sent {severity} damage event for order: {order_id}")
        print(f"📄 Message: {json.dumps(mqtt_message, indent=2)}")
        
        connection.close()
        
    except Exception as e:
        print(f"❌ Error sending message: {e}")
        sys.exit(1)

def main():
    """Main function to handle command line arguments"""
    import argparse
    
    parser = argparse.ArgumentParser(description='Send order damage events to RabbitMQ')
    parser.add_argument('--severity', choices=['minor', 'major', 'critical'], 
                       default='minor', help='Damage severity level')
    parser.add_argument('--order-id', help='Order ID (auto-generated if not provided)')
    parser.add_argument('--count', type=int, default=1, help='Number of events to send')
    
    args = parser.parse_args()
    
    print(f"🚀 Sending {args.count} {args.severity} damage event(s) to RabbitMQ...")
    
    for i in range(args.count):
        order_id = args.order_id
        if args.count > 1 and order_id:
            order_id = f"{args.order_id}_{i+1}"
        
        send_damage_event(args.severity, order_id)
        
        if i < args.count - 1:
            import time
            time.sleep(1)  # Small delay between messages
    
    print(f"🎉 Successfully sent {args.count} event(s)!")

if __name__ == "__main__":
    main()