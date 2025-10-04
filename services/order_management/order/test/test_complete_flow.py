#!/usr/bin/env python3
"""
Complete flow test script:
1. Sends damage events to order-damage-queue
2. Monitors order events from order-events-queue
3. Verifies the complete flow works correctly

Requires: pip install pika
"""

import pika
import json
import sys
import time
import threading
from datetime import datetime, timezone

class OrderFlowTester:
    def __init__(self):
        self.connection = None
        self.channel = None
        self.monitoring = False
        self.received_events = []
        
    def connect(self):
        """Connect to RabbitMQ"""
        try:
            self.connection = pika.BlockingConnection(
                pika.ConnectionParameters(host='localhost', port=5672, 
                                        credentials=pika.PlainCredentials('guest', 'guest'))
            )
            self.channel = self.connection.channel()
            return True
        except Exception as e:
            print(f"❌ Failed to connect to RabbitMQ: {e}")
            return False
    
    def send_damage_event(self, severity="minor", order_id=None):
        """Send a damage event"""
        if not order_id:
            order_id = f"test_flow_{int(datetime.now().timestamp())}"
        
        # Create the damage event payload
        damage_event_payload = {
            "eventId": f"evt_{int(datetime.now().timestamp())}",
            "type": "order.damage",
            "source": "flow-test-script",
            "occurredAt": datetime.now(timezone.utc).isoformat(),
            "orderId": order_id,
            "severity": severity,
            "description": f"Flow test damage event: severity={severity}",
            "details": {
                "temperature": 15.0,
                "humidity": 50,
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
            # Check if connection is still open, reconnect if needed
            if self.connection is None or self.connection.is_closed:
                if not self.connect():
                    return None
            
            # Publish the damage event
            self.channel.basic_publish(
                exchange='events',
                routing_key='order.damage',
                body=json.dumps(mqtt_message),
                properties=pika.BasicProperties(
                    delivery_mode=2,
                    content_type='application/json'
                )
            )
            
            print(f"📤 Sent {severity} damage event for order: {order_id}")
            return order_id
            
        except Exception as e:
            print(f"❌ Error sending damage event: {e}")
            # Try to reconnect and retry once
            if self.connect():
                try:
                    self.channel.basic_publish(
                        exchange='events',
                        routing_key='order.damage',
                        body=json.dumps(mqtt_message),
                        properties=pika.BasicProperties(
                            delivery_mode=2,
                            content_type='application/json'
                        )
                    )
                    print(f"📤 Sent {severity} damage event for order: {order_id} (after reconnect)")
                    return order_id
                except Exception as retry_e:
                    print(f"❌ Error sending damage event after reconnect: {retry_e}")
            return None
    
    def monitor_order_events(self, timeout_seconds=30):
        """Monitor order events for a specified time"""
        self.monitoring = True
        self.received_events = []
        
        # Ensure we have a fresh connection for monitoring
        if self.connection is None or self.connection.is_closed:
            if not self.connect():
                return
        
        def callback(ch, method, properties, body):
            try:
                event = json.loads(body.decode('utf-8'))
                self.received_events.append(event)
                
                print(f"📨 Received Order Event:")
                print(f"   Event Type: {event.get('event_type', 'N/A')}")
                print(f"   Order ID: {event.get('order_id', 'N/A')}")
                
                if 'order' in event:
                    order = event['order']
                    print(f"   Order Status: {order.get('status', 'N/A')}")
                
                print("-" * 40)
                ch.basic_ack(delivery_tag=method.delivery_tag)
                
            except Exception as e:
                print(f"❌ Error processing order event: {e}")
                ch.basic_ack(delivery_tag=method.delivery_tag)
        
        try:
            # Set up consumer
            self.channel.basic_qos(prefetch_count=1)
            self.channel.basic_consume(queue='order-events-queue', on_message_callback=callback)
            
            # Start consuming in a separate thread
            def consume():
                try:
                    self.channel.start_consuming()
                except Exception as e:
                    print(f"❌ Error in consumer thread: {e}")
            
            consumer_thread = threading.Thread(target=consume)
            consumer_thread.daemon = True
            consumer_thread.start()
            
            # Wait for timeout
            time.sleep(timeout_seconds)
            
            # Stop monitoring
            self.monitoring = False
            try:
                self.channel.stop_consuming()
            except:
                pass
                
        except Exception as e:
            print(f"❌ Error setting up monitoring: {e}")
    
    def run_test_scenario(self, scenario_name, damage_events, monitor_time=10):
        """Run a complete test scenario"""
        print(f"\n🧪 Running Test Scenario: {scenario_name}")
        print("=" * 60)
        
        # Ensure fresh connection for this scenario
        self.close()
        if not self.connect():
            print("❌ Failed to connect for this scenario")
            return False
        
        # Send damage events
        sent_orders = []
        for severity, order_id in damage_events:
            sent_order = self.send_damage_event(severity, order_id)
            if sent_order:
                sent_orders.append(sent_order)
            time.sleep(1)  # Small delay between events
        
        print(f"\n🔍 Monitoring order events for {monitor_time} seconds...")
        
        # Monitor for order events
        self.monitor_order_events(monitor_time)
        
        # Analyze results
        print(f"\n📊 Test Results:")
        print(f"   Damage events sent: {len(sent_orders)}")
        print(f"   Order events received: {len(self.received_events)}")
        
        if self.received_events:
            print(f"\n📋 Received Events Summary:")
            for i, event in enumerate(self.received_events, 1):
                event_type = event.get('event_type', 'unknown')
                order_id = event.get('order_id', 'unknown')
                status = 'unknown'
                if 'order' in event:
                    status = event['order'].get('status', 'unknown')
                print(f"   {i}. {event_type} - Order: {order_id} - Status: {status}")
        
        return len(self.received_events) > 0
    
    def close(self):
        """Close connection"""
        try:
            if self.channel and not self.channel.is_closed:
                self.channel.close()
        except:
            pass
        try:
            if self.connection and not self.connection.is_closed:
                self.connection.close()
        except:
            pass
        self.connection = None
        self.channel = None

def main():
    """Main test function"""
    print("🚀 Starting Complete Order Flow Test")
    print("=" * 60)
    
    tester = OrderFlowTester()
    
    if not tester.connect():
        sys.exit(1)
    
    try:
        # Test Scenario 1: Single minor damage
        success1 = tester.run_test_scenario(
            "Single Minor Damage",
            [("minor", "FLOW_TEST_001")],
            monitor_time=8
        )
        
        # Test Scenario 2: Multiple severities
        success2 = tester.run_test_scenario(
            "Multiple Severity Levels",
            [
                ("minor", "FLOW_TEST_002"),
                ("major", "FLOW_TEST_003"),
                ("critical", "FLOW_TEST_004")
            ],
            monitor_time=12
        )
        
        # Test Scenario 3: Same order, escalating damage
        success3 = tester.run_test_scenario(
            "Escalating Damage on Same Order",
            [
                ("minor", "FLOW_TEST_005"),
                ("major", "FLOW_TEST_005"),  # Same order ID
                ("critical", "FLOW_TEST_005")  # Same order ID
            ],
            monitor_time=15
        )
        
        # Final results
        print(f"\n🎯 Final Test Results:")
        print(f"   Scenario 1 (Single Minor): {'✅ PASS' if success1 else '❌ FAIL'}")
        print(f"   Scenario 2 (Multiple Severities): {'✅ PASS' if success2 else '❌ FAIL'}")
        print(f"   Scenario 3 (Escalating Damage): {'✅ PASS' if success3 else '❌ FAIL'}")
        
        overall_success = success1 and success2 and success3
        print(f"\n🏆 Overall Test Result: {'✅ ALL TESTS PASSED' if overall_success else '❌ SOME TESTS FAILED'}")
        
        if not overall_success:
            print("\n💡 Troubleshooting tips:")
            print("   - Check if the order service is running: docker logs order-management")
            print("   - Verify RabbitMQ queues: http://localhost:15672")
            print("   - Check queue bindings and message routing")
        
    except KeyboardInterrupt:
        print("\n🛑 Test interrupted by user")
    except Exception as e:
        print(f"\n❌ Test failed with error: {e}")
    finally:
        tester.close()

if __name__ == "__main__":
    main()