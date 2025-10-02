#!/usr/bin/env python3
"""
Kafka-RabbitMQ Replicator - Bidirectional message replication
"""

import json
import time
import logging
import os
import sys
import traceback
import signal
from typing import Dict, Set, Optional, List
from kafka import KafkaConsumer, KafkaProducer, TopicPartition
from kafka.errors import KafkaError, NoBrokersAvailable
import pika
from pika.exceptions import AMQPConnectionError, AMQPChannelError

class KafkaRabbitMQReplicator:
    def __init__(self, direction="K2R"):
        self.direction = direction  # K2R (Kafka to RabbitMQ) or R2K (RabbitMQ to Kafka)
        self.setup_logging()
        self.load_config()
        self.message_count = 0
        self.error_count = 0
        self.last_message_time = None
        self.start_time = time.time()
        self.heartbeat_interval = 60
        self.last_heartbeat = time.time()
        self.processed_messages: Set[str] = set()
        self.shutdown_requested = False
        
        # Connections
        self.kafka_consumer: Optional[KafkaConsumer] = None
        self.kafka_producer: Optional[KafkaProducer] = None
        self.rabbitmq_connection: Optional[pika.BlockingConnection] = None
        self.rabbitmq_channel: Optional[pika.channel.Channel] = None
        
        # Setup graceful shutdown
        signal.signal(signal.SIGINT, self._signal_handler)
        signal.signal(signal.SIGTERM, self._signal_handler)
        
    def setup_logging(self):
        """Setup structured logging"""
        logging.basicConfig(
            level=logging.INFO,
            format=f'%(asctime)s [{self.direction}] %(levelname)s: %(message)s',
            datefmt='%Y-%m-%d %H:%M:%S'
        )
        self.logger = logging.getLogger(__name__)
        
    def _signal_handler(self, signum, frame):
        """Handle shutdown signals gracefully"""
        self.logger.info(f"Received signal {signum}, initiating graceful shutdown...")
        self.shutdown_requested = True
        
    def load_config(self):
        """Load and validate configuration from environment"""
        self.logger.info(f"=== {self.direction} REPLICATOR STARTING ===")
        
        # Load environment variables
        self.kafka_servers = os.getenv('KAFKA_BOOTSTRAP_SERVERS', 'kafka:9092').split(',')
        self.rabbitmq_host = os.getenv('RABBITMQ_HOST', 'rabbitmq')
        self.rabbitmq_port = int(os.getenv('RABBITMQ_PORT', '5672'))
        self.rabbitmq_username = os.getenv('RABBITMQ_USERNAME', 'user')
        self.rabbitmq_password = os.getenv('RABBITMQ_PASSWORD', 'password')
        self.rabbitmq_vhost = os.getenv('RABBITMQ_VHOST', '/')
        
        # Consumer group
        if self.direction == "K2R":
            self.consumer_group = os.getenv('CONSUMER_GROUP', 'kafka-rabbitmq-replicator-k2r')
        else:
            self.consumer_group = os.getenv('CONSUMER_GROUP', 'kafka-rabbitmq-replicator-r2k')
        
        # Load mappings
        mappings_str = os.getenv('REPLICATION_MAPPINGS', '[]')
        
        self.logger.info(f"KAFKA_BOOTSTRAP_SERVERS: {self.kafka_servers}")
        self.logger.info(f"RABBITMQ_HOST: {self.rabbitmq_host}:{self.rabbitmq_port}")
        self.logger.info(f"RABBITMQ_VHOST: {self.rabbitmq_vhost}")
        self.logger.info(f"CONSUMER_GROUP: {self.consumer_group}")
        self.logger.info(f"REPLICATION_MAPPINGS: {mappings_str}")
        
        # Parse mappings
        try:
            self.mappings = json.loads(mappings_str)
            if not isinstance(self.mappings, list):
                raise ValueError("Mappings must be a JSON array")
        except (json.JSONDecodeError, ValueError) as e:
            self.logger.error(f"Invalid REPLICATION_MAPPINGS: {e}")
            sys.exit(1)
            
        if not self.mappings:
            self.logger.error("No mappings configured for replication")
            sys.exit(1)
            
        self.logger.info(f"Loaded {len(self.mappings)} replication mappings")
        for i, mapping in enumerate(self.mappings):
            self.logger.info(f"Mapping {i+1}: {mapping}")
            
    def create_kafka_consumer(self):
        """Create Kafka consumer for K2R direction"""
        if self.direction != "K2R":
            return None
            
        self.logger.info("Creating Kafka consumer...")
        
        # Extract Kafka topics from mappings
        kafka_topics = [mapping['kafkaTopic'] for mapping in self.mappings]
        
        max_retries = 5
        retry_delay = 5
        
        for attempt in range(max_retries):
            try:
                self.logger.info(f"Kafka consumer creation attempt {attempt + 1}/{max_retries}")
                
                consumer = KafkaConsumer(
                    *kafka_topics,
                    bootstrap_servers=self.kafka_servers,
                    group_id=self.consumer_group,
                    
                    # Deserialization
                    value_deserializer=lambda m: m if m is None else m,
                    key_deserializer=lambda m: m if m is None else m,
                    
                    # Offset management
                    auto_offset_reset='latest',
                    enable_auto_commit=True,
                    auto_commit_interval_ms=1000,
                    
                    # Timeouts
                    consumer_timeout_ms=None,
                    session_timeout_ms=30000,
                    heartbeat_interval_ms=3000,
                    request_timeout_ms=40000,
                    
                    # Fetching
                    max_poll_records=100,
                    max_poll_interval_ms=300000,
                    fetch_min_bytes=1,
                    fetch_max_wait_ms=1000,
                )
                
                self.logger.info("✅ Kafka consumer created successfully")
                return consumer
                
            except Exception as e:
                self.logger.error(f"❌ Kafka consumer creation failed: {e}")
                if attempt < max_retries - 1:
                    self.logger.info(f"Retrying in {retry_delay} seconds...")
                    time.sleep(retry_delay)
                else:
                    raise
                    
        return None
        
    def create_kafka_producer(self):
        """Create Kafka producer for R2K direction"""
        if self.direction != "R2K":
            return None
            
        self.logger.info("Creating Kafka producer...")
        
        max_retries = 5
        retry_delay = 5
        
        for attempt in range(max_retries):
            try:
                self.logger.info(f"Kafka producer creation attempt {attempt + 1}/{max_retries}")
                
                producer = KafkaProducer(
                    bootstrap_servers=self.kafka_servers,
                    
                    # Serialization
                    value_serializer=lambda v: v if isinstance(v, bytes) else v.encode('utf-8') if v else None,
                    key_serializer=lambda k: k if isinstance(k, bytes) else k.encode('utf-8') if k else None,
                    
                    # Reliability
                    acks='all',
                    retries=5,
                    retry_backoff_ms=1000,
                    max_in_flight_requests_per_connection=1,
                    
                    # Timeouts
                    request_timeout_ms=30000,
                    delivery_timeout_ms=120000,
                    
                    # Batching
                    batch_size=16384,
                    linger_ms=5,
                    
                    # Buffer
                    buffer_memory=33554432,
                    
                    # Compression
                    compression_type='gzip',
                    
                    # Idempotence
                    enable_idempotence=True,
                )
                
                self.logger.info("✅ Kafka producer created successfully")
                return producer
                
            except Exception as e:
                self.logger.error(f"❌ Kafka producer creation failed: {e}")
                if attempt < max_retries - 1:
                    self.logger.info(f"Retrying in {retry_delay} seconds...")
                    time.sleep(retry_delay)
                else:
                    raise
                    
        return None
        
    def create_rabbitmq_connection(self):
        """Create RabbitMQ connection and channel"""
        self.logger.info("Creating RabbitMQ connection...")
        
        max_retries = 5
        retry_delay = 5
        
        for attempt in range(max_retries):
            try:
                self.logger.info(f"RabbitMQ connection attempt {attempt + 1}/{max_retries}")
                
                # Connection parameters
                credentials = pika.PlainCredentials(self.rabbitmq_username, self.rabbitmq_password)
                parameters = pika.ConnectionParameters(
                    host=self.rabbitmq_host,
                    port=self.rabbitmq_port,
                    virtual_host=self.rabbitmq_vhost,
                    credentials=credentials,
                    heartbeat=600,
                    blocked_connection_timeout=300,
                )
                
                # Create connection
                connection = pika.BlockingConnection(parameters)
                channel = connection.channel()
                
                self.logger.info("✅ RabbitMQ connection created successfully")
                
                # Declare exchanges and queues based on mappings
                self._setup_rabbitmq_topology(channel)
                
                return connection, channel
                
            except Exception as e:
                self.logger.error(f"❌ RabbitMQ connection failed: {e}")
                if attempt < max_retries - 1:
                    self.logger.info(f"Retrying in {retry_delay} seconds...")
                    time.sleep(retry_delay)
                else:
                    raise
                    
        return None, None
        
    def _setup_rabbitmq_topology(self, channel):
        """Setup RabbitMQ exchanges and queues"""
        self.logger.info("Setting up RabbitMQ topology...")
        
        exchanges = set()
        queues = set()
        
        for mapping in self.mappings:
            if self.direction == "K2R":
                # Kafka to RabbitMQ mappings
                exchange = mapping.get('rabbitmqExchange')
                queue = mapping.get('rabbitmqQueue')
                routing_key = mapping.get('rabbitmqRoutingKey', '')
                
                if exchange:
                    exchanges.add(exchange)
                if queue:
                    queues.add((queue, exchange, routing_key))
                    
            else:  # R2K
                # RabbitMQ to Kafka mappings
                queue = mapping.get('rabbitmqQueue')
                if queue:
                    queues.add((queue, None, ''))
        
        # Declare exchanges
        for exchange in exchanges:
            try:
                channel.exchange_declare(exchange=exchange, exchange_type='topic', durable=True)
                self.logger.info(f"✅ Declared exchange: {exchange}")
            except Exception as e:
                self.logger.error(f"❌ Failed to declare exchange {exchange}: {e}")
        
        # Declare queues and bindings
        for queue_info in queues:
            queue, exchange, routing_key = queue_info
            try:
                channel.queue_declare(queue=queue, durable=True)
                self.logger.info(f"✅ Declared queue: {queue}")
                
                if exchange and routing_key:
                    channel.queue_bind(exchange=exchange, queue=queue, routing_key=routing_key)
                    self.logger.info(f"✅ Bound queue {queue} to exchange {exchange} with routing key {routing_key}")
                    
            except Exception as e:
                self.logger.error(f"❌ Failed to setup queue {queue}: {e}")    def 
process_kafka_message(self, message):
        """Process message from Kafka to RabbitMQ"""
        try:
            # Find mapping for this topic
            mapping = None
            for m in self.mappings:
                if m['kafkaTopic'] == message.topic:
                    mapping = m
                    break
                    
            if not mapping:
                self.logger.warning(f"No mapping found for topic: {message.topic}")
                return False
            
            # Create unique message ID for deduplication
            message_id = f"k2r:{message.topic}:{message.partition}:{message.offset}"
            
            if message_id in self.processed_messages:
                return True
            
            self.message_count += 1
            self.last_message_time = time.time()
            
            # Decode message
            try:
                value = message.value
                key = message.key
                
                if isinstance(value, bytes):
                    try:
                        value = value.decode('utf-8')
                    except UnicodeDecodeError:
                        pass
                        
                if isinstance(key, bytes):
                    try:
                        key = key.decode('utf-8')
                    except UnicodeDecodeError:
                        pass
                        
            except Exception as decode_error:
                self.logger.warning(f"Message decode warning: {decode_error}")
                value = message.value
                key = message.key
            
            # Log processing
            if self.message_count % 100 == 0 or self.message_count <= 10:
                self.logger.info(
                    f"[MSG #{self.message_count}] Processing Kafka->RabbitMQ: "
                    f"{message.topic}:{message.partition}:{message.offset} -> "
                    f"{mapping.get('rabbitmqQueue', 'N/A')}"
                )
            
            # Publish to RabbitMQ
            exchange = mapping.get('rabbitmqExchange', '')
            queue = mapping.get('rabbitmqQueue', '')
            routing_key = mapping.get('rabbitmqRoutingKey', '')
            
            # Prepare message body
            if isinstance(value, str):
                body = value.encode('utf-8')
            else:
                body = value or b''
            
            # Prepare properties
            properties = pika.BasicProperties(
                delivery_mode=2,  # Make message persistent
                headers={
                    'kafka_topic': message.topic,
                    'kafka_partition': message.partition,
                    'kafka_offset': message.offset,
                    'kafka_key': key,
                    'replicator_id': message_id
                }
            )
            
            # Publish message
            if exchange:
                self.rabbitmq_channel.basic_publish(
                    exchange=exchange,
                    routing_key=routing_key,
                    body=body,
                    properties=properties
                )
            else:
                # Direct queue publish
                self.rabbitmq_channel.basic_publish(
                    exchange='',
                    routing_key=queue,
                    body=body,
                    properties=properties
                )
            
            # Track processed message
            self.processed_messages.add(message_id)
            
            # Limit memory usage
            if len(self.processed_messages) > 10000:
                old_messages = list(self.processed_messages)[:1000]
                for old_msg in old_messages:
                    self.processed_messages.discard(old_msg)
            
            return True
            
        except Exception as e:
            self.error_count += 1
            self.logger.error(
                f"[MSG #{self.message_count}] Failed to process Kafka message "
                f"{message.topic}:{message.partition}:{message.offset}: {e}"
            )
            return False
            
    def process_rabbitmq_message(self, channel, method, properties, body):
        """Process message from RabbitMQ to Kafka"""
        try:
            # Find mapping for this queue
            queue_name = method.routing_key if method.exchange == '' else None
            if not queue_name:
                # Try to get queue name from method
                queue_name = getattr(method, 'consumer_tag', 'unknown')
            
            mapping = None
            for m in self.mappings:
                if m.get('rabbitmqQueue') == queue_name:
                    mapping = m
                    break
                    
            if not mapping:
                self.logger.warning(f"No mapping found for queue: {queue_name}")
                channel.basic_nack(delivery_tag=method.delivery_tag, requeue=False)
                return
            
            # Create unique message ID for deduplication
            message_id = f"r2k:{queue_name}:{method.delivery_tag}"
            
            if message_id in self.processed_messages:
                channel.basic_ack(delivery_tag=method.delivery_tag)
                return
            
            self.message_count += 1
            self.last_message_time = time.time()
            
            # Decode message
            try:
                if isinstance(body, bytes):
                    try:
                        value = body.decode('utf-8')
                    except UnicodeDecodeError:
                        value = body
                else:
                    value = body
                    
            except Exception as decode_error:
                self.logger.warning(f"Message decode warning: {decode_error}")
                value = body
            
            # Extract key from headers if available
            key = None
            if properties and properties.headers:
                key = properties.headers.get('kafka_key')
            
            # Log processing
            if self.message_count % 100 == 0 or self.message_count <= 10:
                self.logger.info(
                    f"[MSG #{self.message_count}] Processing RabbitMQ->Kafka: "
                    f"{queue_name} -> {mapping['kafkaTopic']}"
                )
            
            # Send to Kafka
            kafka_topic = mapping['kafkaTopic']
            
            future = self.kafka_producer.send(
                kafka_topic,
                key=key,
                value=value,
                headers=[
                    ('rabbitmq_queue', queue_name.encode('utf-8') if queue_name else b''),
                    ('rabbitmq_exchange', (method.exchange or '').encode('utf-8')),
                    ('rabbitmq_routing_key', (method.routing_key or '').encode('utf-8')),
                    ('replicator_id', message_id.encode('utf-8'))
                ]
            )
            
            # Wait for completion
            record_metadata = future.get(timeout=30)
            
            # Acknowledge RabbitMQ message
            channel.basic_ack(delivery_tag=method.delivery_tag)
            
            # Track processed message
            self.processed_messages.add(message_id)
            
            # Limit memory usage
            if len(self.processed_messages) > 10000:
                old_messages = list(self.processed_messages)[:1000]
                for old_msg in old_messages:
                    self.processed_messages.discard(old_msg)
            
        except Exception as e:
            self.error_count += 1
            self.logger.error(f"[MSG #{self.message_count}] Failed to process RabbitMQ message: {e}")
            # Nack and don't requeue to avoid infinite loops
            channel.basic_nack(delivery_tag=method.delivery_tag, requeue=False)
            
    def log_heartbeat(self):
        """Log periodic heartbeat with stats"""
        uptime = time.time() - self.start_time
        
        self.logger.info("=" * 50)
        self.logger.info(f"HEARTBEAT - Uptime: {uptime:.0f}s")
        self.logger.info(f"Messages processed: {self.message_count}")
        self.logger.info(f"Errors: {self.error_count}")
        
        if self.last_message_time:
            time_since_last = time.time() - self.last_message_time
            self.logger.info(f"Last message: {time_since_last:.1f}s ago")
        else:
            self.logger.info("Last message: Never")
            
        if self.message_count > 0:
            rate = self.message_count / uptime
            self.logger.info(f"Average rate: {rate:.2f} msg/s")
        
        self.logger.info("=" * 50)
        
    def run_kafka_to_rabbitmq(self):
        """Main loop for Kafka to RabbitMQ replication"""
        self.logger.info("Starting Kafka to RabbitMQ replication...")
        
        # Create connections
        self.kafka_consumer = self.create_kafka_consumer()
        if not self.kafka_consumer:
            self.logger.error("Failed to create Kafka consumer")
            sys.exit(1)
            
        self.rabbitmq_connection, self.rabbitmq_channel = self.create_rabbitmq_connection()
        if not self.rabbitmq_connection or not self.rabbitmq_channel:
            self.logger.error("Failed to create RabbitMQ connection")
            sys.exit(1)
        
        self.logger.info("=== K2R REPLICATION STARTED ===")
        self.logger.info("Listening for Kafka messages...")
        
        consecutive_empty_polls = 0
        
        try:
            while not self.shutdown_requested:
                current_time = time.time()
                
                # Heartbeat
                if current_time - self.last_heartbeat >= self.heartbeat_interval:
                    self.log_heartbeat()
                    self.last_heartbeat = current_time
                
                try:
                    # Poll for messages
                    if consecutive_empty_polls % 10 == 0:
                        self.logger.info(f"🔍 Polling for Kafka messages... (poll #{consecutive_empty_polls + 1})")
                    
                    message_batch = self.kafka_consumer.poll(timeout_ms=5000)
                    
                    if message_batch:
                        consecutive_empty_polls = 0
                        total_messages = sum(len(messages) for messages in message_batch.values())
                        self.logger.info(f"📨 Received {total_messages} Kafka messages!")
                        
                        # Process messages
                        batch_count = 0
                        for topic_partition, messages in message_batch.items():
                            for message in messages:
                                if self.shutdown_requested:
                                    break
                                    
                                success = self.process_kafka_message(message)
                                if success:
                                    batch_count += 1
                        
                        if batch_count > 0:
                            self.logger.info(f"✅ Processed batch of {batch_count} messages")
                            
                    else:
                        consecutive_empty_polls += 1
                        
                        if consecutive_empty_polls <= 5 or consecutive_empty_polls % 30 == 0:
                            self.logger.info(f"No Kafka messages in poll #{consecutive_empty_polls}")
                
                except Exception as e:
                    self.logger.error(f"Error in K2R polling loop: {e}")
                    self.error_count += 1
                    time.sleep(1)
                    
                time.sleep(0.01)
                
        except KeyboardInterrupt:
            self.logger.info("Shutdown requested via KeyboardInterrupt")
            self.shutdown_requested = True
            
    def run_rabbitmq_to_kafka(self):
        """Main loop for RabbitMQ to Kafka replication"""
        self.logger.info("Starting RabbitMQ to Kafka replication...")
        
        # Create connections
        self.kafka_producer = self.create_kafka_producer()
        if not self.kafka_producer:
            self.logger.error("Failed to create Kafka producer")
            sys.exit(1)
            
        self.rabbitmq_connection, self.rabbitmq_channel = self.create_rabbitmq_connection()
        if not self.rabbitmq_connection or not self.rabbitmq_channel:
            self.logger.error("Failed to create RabbitMQ connection")
            sys.exit(1)
        
        self.logger.info("=== R2K REPLICATION STARTED ===")
        
        # Setup consumers for each queue
        for mapping in self.mappings:
            queue = mapping.get('rabbitmqQueue')
            if queue:
                self.logger.info(f"Setting up consumer for queue: {queue}")
                self.rabbitmq_channel.basic_consume(
                    queue=queue,
                    on_message_callback=self.process_rabbitmq_message,
                    auto_ack=False
                )
        
        self.logger.info("Listening for RabbitMQ messages...")
        
        try:
            # Start consuming with timeout
            while not self.shutdown_requested:
                current_time = time.time()
                
                # Heartbeat
                if current_time - self.last_heartbeat >= self.heartbeat_interval:
                    self.log_heartbeat()
                    self.last_heartbeat = current_time
                
                try:
                    # Process RabbitMQ messages with timeout
                    self.rabbitmq_connection.process_data_events(time_limit=1)
                    
                except Exception as e:
                    self.logger.error(f"Error in R2K processing loop: {e}")
                    self.error_count += 1
                    time.sleep(1)
                    
        except KeyboardInterrupt:
            self.logger.info("Shutdown requested via KeyboardInterrupt")
            self.shutdown_requested = True
            
    def run(self):
        """Main entry point"""
        try:
            if self.direction == "K2R":
                self.run_kafka_to_rabbitmq()
            elif self.direction == "R2K":
                self.run_rabbitmq_to_kafka()
            else:
                self.logger.error(f"Invalid direction: {self.direction}")
                sys.exit(1)
                
        except Exception as e:
            self.logger.error(f"Fatal error in main loop: {e}")
            self.logger.error(f"Traceback: {traceback.format_exc()}")
            sys.exit(1)
            
        finally:
            self._cleanup()
            
    def _cleanup(self):
        """Clean up resources"""
        self.logger.info("Cleaning up resources...")
        
        try:
            if self.kafka_producer:
                self.logger.info("Flushing Kafka producer...")
                self.kafka_producer.flush(timeout=10)
                self.kafka_producer.close()
                
            if self.kafka_consumer:
                self.logger.info("Closing Kafka consumer...")
                self.kafka_consumer.close()
                
            if self.rabbitmq_connection and not self.rabbitmq_connection.is_closed:
                self.logger.info("Closing RabbitMQ connection...")
                self.rabbitmq_connection.close()
                
        except Exception as e:
            self.logger.error(f"Error during cleanup: {e}")
            
        self.logger.info(f"Final stats: {self.message_count} messages processed, {self.error_count} errors")
        self.logger.info("Replicator shutdown complete")

if __name__ == "__main__":
    direction = sys.argv[1] if len(sys.argv) > 1 else "K2R"
    replicator = KafkaRabbitMQReplicator(direction)
    replicator.run()