#!/usr/bin/env python3
"""
Test script to monitor order events from the order-events-queue
Requires: pip install pika
"""

import pika
import json
import sys
from datetime import datetime

def monitor_order_events():
    """Monitor order events from the order-events-queue"""
    
    try:
        # Connect to RabbitMQ
        connection = pika.BlockingConnection(
            pika.ConnectionParameters(host='localhost', port=5672, 
                                    credentials=pika.PlainCredentials('guest', 'guest'))
        )
        channel = connection.channel()
        
        # Ensure the queue exists
        channel.queue_declare(queue='order-events-queue', durable=True)
        
        print("🔍 Monitoring order events from 'order-events-queue'...")
        print("📋 Waiting for messages. To exit press CTRL+C")
        print("-" * 60)
        
        def callback(ch, method, properties, body):
            """Callback function to process received messages"""
            try:
                # Parse the JSON message
                event = json.loads(body.decode('utf-8'))
                
                print(f"📨 Received Order Event:")
                print(f"   Event Type: {event.get('event_type', 'N/A')}")
                print(f"   Order ID: {event.get('order_id', 'N/A')}")
                print(f"   Timestamp: {event.get('timestamp', 'N/A')}")
                
                if 'order' in event:
                    order = event['order']
                    print(f"   Order Status: {order.get('status', 'N/A')}")
                    print(f"   Customer ID: {order.get('customer_id', 'N/A')}")
                    print(f"   Product ID: {order.get('product_id', 'N/A')}")
                    print(f"   Total Amount: ${order.get('total_amount', 0):.2f}")
                
                print(f"   Raw Message: {json.dumps(event, indent=2)}")
                print("-" * 60)
                
                # Acknowledge the message
                ch.basic_ack(delivery_tag=method.delivery_tag)
                
            except json.JSONDecodeError as e:
                print(f"❌ Error parsing JSON: {e}")
                print(f"   Raw body: {body.decode('utf-8')}")
                print("-" * 60)
                # Still acknowledge to avoid reprocessing
                ch.basic_ack(delivery_tag=method.delivery_tag)
            except Exception as e:
                print(f"❌ Error processing message: {e}")
                print("-" * 60)
                # Still acknowledge to avoid reprocessing
                ch.basic_ack(delivery_tag=method.delivery_tag)
        
        # Set up consumer
        channel.basic_qos(prefetch_count=1)
        channel.basic_consume(queue='order-events-queue', on_message_callback=callback)
        
        # Start consuming
        channel.start_consuming()
        
    except KeyboardInterrupt:
        print("\n🛑 Stopping monitor...")
        channel.stop_consuming()
        connection.close()
        print("✅ Monitor stopped successfully!")
        
    except Exception as e:
        print(f"❌ Error monitoring events: {e}")
        sys.exit(1)

def check_queue_status():
    """Check the status of the order-events-queue"""
    try:
        # Connect to RabbitMQ
        connection = pika.BlockingConnection(
            pika.ConnectionParameters(host='localhost', port=5672, 
                                    credentials=pika.PlainCredentials('guest', 'guest'))
        )
        channel = connection.channel()
        
        # Get queue info
        method = channel.queue_declare(queue='order-events-queue', durable=True, passive=True)
        message_count = method.method.message_count
        consumer_count = method.method.consumer_count
        
        print(f"📊 Queue Status: order-events-queue")
        print(f"   Messages: {message_count}")
        print(f"   Consumers: {consumer_count}")
        
        connection.close()
        
    except Exception as e:
        print(f"❌ Error checking queue status: {e}")

def main():
    """Main function to handle command line arguments"""
    import argparse
    
    parser = argparse.ArgumentParser(description='Monitor order events from RabbitMQ')
    parser.add_argument('--status', action='store_true', help='Check queue status only')
    
    args = parser.parse_args()
    
    if args.status:
        check_queue_status()
    else:
        monitor_order_events()

if __name__ == "__main__":
    main()