#!/usr/bin/env python3
"""
Script to send test order events to Kafka for testing the warehouse batch service
"""

import json
import sys
from datetime import datetime, timezone
from kafka import KafkaProducer
from kafka.errors import KafkaError

def create_test_order_event(event_type="order.damage_processed", order_id=None):
    """Create a test order event"""
    if order_id is None:
        order_id = f"test_{int(datetime.now().timestamp())}"
    
    now = datetime.now(timezone.utc).isoformat()
    
    event = {
        "event_type": event_type,
        "order_id": order_id,
        "order": {
            "id": order_id,
            "customer_id": "test_customer_123",
            "product_id": "test_product_456",
            "quantity": 2,
            "status": "damage_detected_minor" if "damage" in event_type else "pending",
            "total_amount": 99.99,
            "created_at": now,
            "updated_at": now
        },
        "timestamp": now
    }
    
    return event

def send_event_to_kafka(event, topic="order-events", bootstrap_servers="localhost:9092"):
    """Send event to Kafka"""
    try:
        producer = KafkaProducer(
            bootstrap_servers=[bootstrap_servers],
            value_serializer=lambda v: json.dumps(v).encode('utf-8'),
            key_serializer=lambda k: k.encode('utf-8') if k else None
        )
        
        # Send the event
        future = producer.send(
            topic, 
            value=event, 
            key=event['order_id']
        )
        
        # Wait for the message to be sent
        record_metadata = future.get(timeout=10)
        
        print(f"✅ Event sent successfully!")
        print(f"   Topic: {record_metadata.topic}")
        print(f"   Partition: {record_metadata.partition}")
        print(f"   Offset: {record_metadata.offset}")
        print(f"   Event Type: {event['event_type']}")
        print(f"   Order ID: {event['order_id']}")
        
        producer.close()
        return True
        
    except KafkaError as e:
        print(f"❌ Failed to send event: {e}")
        return False
    except Exception as e:
        print(f"❌ Unexpected error: {e}")
        return False

def main():
    """Main function"""
    import argparse
    
    parser = argparse.ArgumentParser(description="Send test order events to Kafka")
    parser.add_argument("--event-type", default="order.damage_processed",
                       choices=["order.damage_processed", "order.created", "order.cancelled", 
                               "order.shipped", "order.delivered", "order.returned"],
                       help="Type of order event to send")
    parser.add_argument("--order-id", help="Custom order ID (auto-generated if not provided)")
    parser.add_argument("--topic", default="order-events", help="Kafka topic")
    parser.add_argument("--broker", default="localhost:9092", help="Kafka broker address")
    parser.add_argument("--count", type=int, default=1, help="Number of events to send")
    
    args = parser.parse_args()
    
    print(f"🚀 Sending {args.count} test event(s) to Kafka...")
    print(f"   Broker: {args.broker}")
    print(f"   Topic: {args.topic}")
    print(f"   Event Type: {args.event_type}")
    print()
    
    success_count = 0
    for i in range(args.count):
        order_id = args.order_id if args.count == 1 and args.order_id else None
        event = create_test_order_event(args.event_type, order_id)
        
        if send_event_to_kafka(event, args.topic, args.broker):
            success_count += 1
        
        if i < args.count - 1:
            print()
    
    print(f"\n📊 Summary: {success_count}/{args.count} events sent successfully")
    
    if success_count > 0:
        print("\n💡 Check your warehouse batch service logs to see the events being processed!")

if __name__ == "__main__":
    main()