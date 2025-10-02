#!/usr/bin/env python3
"""
Kafka Replicator - Optimized version for reliable message replication
"""

import json
import time
import logging
import os
import sys
import traceback
import signal
from typing import Dict, Set, Optional
from kafka import KafkaConsumer, KafkaProducer, TopicPartition
from kafka.errors import KafkaError, NoBrokersAvailable, CommitFailedError

class KafkaReplicator:
    def __init__(self, direction="S2T"):
        self.direction = direction
        self.setup_logging()
        self.load_config()
        self.message_count = 0
        self.error_count = 0
        self.last_message_time = None
        self.start_time = time.time()
        self.heartbeat_interval = 60  # Reduced frequency
        self.last_heartbeat = time.time()
        self.processed_offsets: Dict[TopicPartition, int] = {}
        self.processed_messages: Set[str] = set()
        self.shutdown_requested = False
        self.consumer: Optional[KafkaConsumer] = None
        self.producer: Optional[KafkaProducer] = None
        
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
        source_servers_str = os.getenv('SOURCE_BOOTSTRAP_SERVERS')
        target_servers_str = os.getenv('TARGET_BOOTSTRAP_SERVERS')
        topic_mapping_str = os.getenv('TOPIC_MAPPING', '{}')
        
        self.logger.info(f"SOURCE_BOOTSTRAP_SERVERS: {source_servers_str}")
        self.logger.info(f"TARGET_BOOTSTRAP_SERVERS: {target_servers_str}")
        self.logger.info(f"TOPIC_MAPPING: {topic_mapping_str}")
        
        # Validate required variables
        if not source_servers_str or not target_servers_str:
            self.logger.error("Missing required environment variables")
            sys.exit(1)
            
        # Parse servers
        self.source_servers = [s.strip() for s in source_servers_str.split(',') if s.strip()]
        self.target_servers = [s.strip() for s in target_servers_str.split(',') if s.strip()]
        
        # Parse topic mapping
        try:
            self.topic_mapping = json.loads(topic_mapping_str)
            if not isinstance(self.topic_mapping, dict):
                raise ValueError("Topic mapping must be a JSON object")
        except (json.JSONDecodeError, ValueError) as e:
            self.logger.error(f"Invalid TOPIC_MAPPING: {e}")
            sys.exit(1)
            
        self.source_topics = list(self.topic_mapping.keys())
        
        if not self.source_topics:
            self.logger.error("No topics configured for replication")
            sys.exit(1)
            
        self.logger.info(f"Topics to replicate: {self.source_topics}")
        self.logger.info(f"Topic mapping: {self.topic_mapping}")
        
    def create_consumer(self):
        """Create consumer without consumer group (manual assignment)"""
        self.logger.info("Creating Kafka consumer...")
        
        max_retries = 5
        retry_delay = 5
        
        for attempt in range(max_retries):
            try:
                self.logger.info(f"Consumer creation attempt {attempt + 1}/{max_retries}")
                
                # Create consumer WITHOUT consumer group
                consumer = KafkaConsumer(
                    bootstrap_servers=self.source_servers,
                    # NO group_id - this is key!
                    
                    # Deserialization - handle bytes properly
                    value_deserializer=lambda m: m if m is None else m,
                    key_deserializer=lambda m: m if m is None else m,
                    
                    # Offset management
                    auto_offset_reset='latest',  # Only new messages by default
                    enable_auto_commit=False,  # No auto commit without group
                    
                    # Timeouts
                    consumer_timeout_ms=None,  # No timeout for polling
                    session_timeout_ms=30000,
                    heartbeat_interval_ms=3000,
                    request_timeout_ms=40000,
                    
                    # Fetching
                    max_poll_records=500,
                    max_poll_interval_ms=300000,
                    fetch_min_bytes=1,
                    fetch_max_wait_ms=1000,
                    
                    # Metadata
                    metadata_max_age_ms=30000,
                    
                    # Connection
                    connections_max_idle_ms=540000,
                    api_version_auto_timeout_ms=60000,
                )
                
                self.logger.info("✅ Consumer created successfully")
                
                # Manual assignment - simple approach like your original
                self.logger.info("Using manual partition assignment...")
                
                partitions = []
                for topic in self.source_topics:
                    # Simple approach: assign partition 0 for each topic
                    # This can be enhanced later to discover all partitions
                    tp = TopicPartition(topic, 0)
                    partitions.append(tp)
                    self.logger.info(f"Will assign: {tp}")
                
                # Direct assignment
                consumer.assign(partitions)
                assignment = consumer.assignment()
                self.logger.info(f"✅ Manual assignment successful: {assignment}")
                
                # Seek to beginning to process all available messages
                self.logger.info("Seeking to beginning to process all messages...")
                consumer.seek_to_beginning()
                
                # Skip detailed position checking for now - it might be hanging
                self.logger.info("Consumer setup complete, starting main loop...")
                
                return consumer
                
            except NoBrokersAvailable:
                self.logger.error(f"❌ No brokers available at {self.source_servers}")
                if attempt < max_retries - 1:
                    self.logger.info(f"Retrying in {retry_delay} seconds...")
                    time.sleep(retry_delay)
                else:
                    raise
                    
            except Exception as e:
                self.logger.error(f"❌ Consumer creation failed: {e}")
                if attempt < max_retries - 1:
                    self.logger.info(f"Retrying in {retry_delay} seconds...")
                    time.sleep(retry_delay)
                else:
                    raise
                    
        return None
    
    def log_consumer_state(self, consumer, label="State"):
        """Log detailed consumer state for debugging"""
        try:
            assignment = consumer.assignment()
            self.logger.info(f"=== {label} ===")
            
            if not assignment:
                self.logger.warning("No partitions assigned")
                return
            
            for tp in assignment:
                try:
                    position = consumer.position(tp)
                    beginning_offsets = consumer.beginning_offsets([tp])
                    end_offsets = consumer.end_offsets([tp])
                    
                    beginning = beginning_offsets.get(tp, -1)
                    end = end_offsets.get(tp, -1)
                    
                    lag = end - position if end >= 0 and position >= 0 else "unknown"
                    available = end - beginning if end >= 0 and beginning >= 0 else "unknown"
                    
                    self.logger.info(
                        f"{tp}: position={position}, "
                        f"beginning={beginning}, end={end}, "
                        f"lag={lag}, available={available}"
                    )
                except Exception as e:
                    self.logger.error(f"Error getting state for {tp}: {e}")
                    
        except Exception as e:
            self.logger.error(f"Error logging consumer state: {e}")
            
    def create_producer(self):
        """Create producer with retry logic"""
        self.logger.info("Creating Kafka producer...")
        
        max_retries = 5
        retry_delay = 5
        
        for attempt in range(max_retries):
            try:
                self.logger.info(f"Producer creation attempt {attempt + 1}/{max_retries}")
                
                producer = KafkaProducer(
                    bootstrap_servers=self.target_servers,
                    
                    # Serialization - handle bytes properly
                    value_serializer=lambda v: v if isinstance(v, bytes) else v.encode('utf-8') if v else None,
                    key_serializer=lambda k: k if isinstance(k, bytes) else k.encode('utf-8') if k else None,
                    
                    # Reliability
                    acks='all',
                    retries=5,
                    retry_backoff_ms=1000,
                    max_in_flight_requests_per_connection=1,  # Ensure ordering
                    
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
                    
                    # Idempotence for exactly-once semantics
                    enable_idempotence=True,
                )
                
                self.logger.info("✅ Producer created successfully")
                return producer
                
            except Exception as e:
                self.logger.error(f"❌ Producer creation failed: {e}")
                if attempt < max_retries - 1:
                    self.logger.info(f"Retrying in {retry_delay} seconds...")
                    time.sleep(retry_delay)
                else:
                    raise
                    
        return None
        
    def process_message(self, message, producer):
        """Process a single message with proper error handling"""
        try:
            source_topic = message.topic
            target_topic = self.topic_mapping.get(source_topic, source_topic)
            
            # Create unique message ID for deduplication
            message_id = f"{source_topic}:{message.partition}:{message.offset}"
            
            # Check if already processed (simple deduplication)
            if message_id in self.processed_messages:
                return True
            
            self.message_count += 1
            self.last_message_time = time.time()
            
            # Decode message value and key if they are bytes
            try:
                value = message.value
                key = message.key
                
                if isinstance(value, bytes):
                    try:
                        value = value.decode('utf-8')
                    except UnicodeDecodeError:
                        # Keep as bytes if can't decode
                        pass
                        
                if isinstance(key, bytes):
                    try:
                        key = key.decode('utf-8')
                    except UnicodeDecodeError:
                        # Keep as bytes if can't decode
                        pass
                        
            except Exception as decode_error:
                self.logger.warning(f"Message decode warning: {decode_error}")
                value = message.value
                key = message.key
            
            # Log every 100 messages to reduce noise
            if self.message_count % 100 == 0 or self.message_count <= 10:
                self.logger.info(
                    f"[MSG #{self.message_count}] Processing: "
                    f"{source_topic}:{message.partition}:{message.offset} -> {target_topic}"
                )
            
            # Send to target
            future = producer.send(
                target_topic,
                key=key,
                value=value,
                headers=message.headers
            )
            
            # Wait for completion
            record_metadata = future.get(timeout=30)
            
            # Track processed message
            source_tp = TopicPartition(source_topic, message.partition)
            self.processed_offsets[source_tp] = message.offset
            self.processed_messages.add(message_id)
            
            # Limit memory usage - keep only last 10000 message IDs
            if len(self.processed_messages) > 10000:
                # Remove oldest entries
                old_messages = list(self.processed_messages)[:1000]
                for old_msg in old_messages:
                    self.processed_messages.discard(old_msg)
            
            return True
            
        except Exception as e:
            self.error_count += 1
            self.logger.error(
                f"[MSG #{self.message_count}] Failed to process message "
                f"{source_topic}:{message.partition}:{message.offset}: {e}"
            )
            return False
            
    def log_heartbeat(self, consumer):
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
        
        # Log consumer lag
        try:
            assignment = consumer.assignment()
            total_lag = 0
            for tp in assignment:
                try:
                    position = consumer.position(tp)
                    end_offsets = consumer.end_offsets([tp])
                    end = end_offsets.get(tp, 0)
                    lag = max(0, end - position) if end >= position else 0
                    total_lag += lag
                    if lag > 0:
                        self.logger.info(f"Lag {tp}: {lag} messages")
                except Exception:
                    pass
            
            if total_lag > 0:
                self.logger.warning(f"Total lag: {total_lag} messages")
            else:
                self.logger.info("No lag detected")
                
        except Exception as e:
            self.logger.warning(f"Could not check lag: {e}")
        
        self.logger.info("=" * 50)
            
    def run(self):
        """Main replication loop with improved error handling"""
        try:
            # Create consumer and producer
            self.consumer = self.create_consumer()
            if not self.consumer:
                self.logger.error("Failed to create consumer")
                sys.exit(1)
                
            self.producer = self.create_producer()
            if not self.producer:
                self.logger.error("Failed to create producer")
                sys.exit(1)
                
            self.logger.info("=== REPLICATION STARTED ===")
            self.logger.info(f"Source topics: {self.source_topics}")
            self.logger.info(f"Topic mapping: {self.topic_mapping}")
            self.logger.info("Using manual assignment (no consumer group)")
            self.logger.info("Listening for messages...")
            
            consecutive_empty_polls = 0
            max_empty_polls = 100  # Prevent infinite empty polling
            
            while not self.shutdown_requested:
                current_time = time.time()
                
                # Heartbeat
                if current_time - self.last_heartbeat >= self.heartbeat_interval:
                    self.log_heartbeat(self.consumer)
                    self.last_heartbeat = current_time
                
                try:
                    # Simple polling without heavy debugging
                    if consecutive_empty_polls % 10 == 0:  # Log every 10 polls
                        self.logger.info(f"🔍 Polling for messages... (poll #{consecutive_empty_polls + 1})")
                    
                    message_batch = self.consumer.poll(timeout_ms=5000)
                    
                    if message_batch:
                        consecutive_empty_polls = 0
                        total_messages = sum(len(messages) for messages in message_batch.values())
                        self.logger.info(f"📨 Received {total_messages} messages!")
                        
                        # Process messages in batches
                        batch_count = 0
                        for topic_partition, messages in message_batch.items():
                            self.logger.info(f"Processing {len(messages)} messages from {topic_partition}")
                            for message in messages:
                                if self.shutdown_requested:
                                    break
                                    
                                success = self.process_message(message, self.producer)
                                if success:
                                    batch_count += 1
                        
                        # Flush producer to ensure delivery
                        self.producer.flush(timeout=10)
                        
                        if batch_count > 0:
                            self.logger.info(f"✅ Processed batch of {batch_count} messages")
                            
                    else:
                        consecutive_empty_polls += 1
                        
                        # Less frequent logging to avoid spam
                        if consecutive_empty_polls <= 5 or consecutive_empty_polls % 30 == 0:
                            self.logger.info(f"No messages in poll #{consecutive_empty_polls}")
                        
                        # Safety check to prevent infinite empty polling
                        if consecutive_empty_polls > max_empty_polls:
                            self.logger.warning("Too many consecutive empty polls, checking consumer health...")
                            
                            # Check if consumer is still healthy
                            try:
                                assignment = self.consumer.assignment()
                                if not assignment:
                                    self.logger.error("Consumer lost partition assignment, recreating...")
                                    self.consumer.close()
                                    self.consumer = self.create_consumer()
                                    consecutive_empty_polls = 0
                            except Exception as e:
                                self.logger.error(f"Consumer health check failed: {e}")
                
                except Exception as e:
                    self.logger.error(f"Error in polling loop: {e}")
                    self.error_count += 1
                    time.sleep(1)  # Brief pause on error
                    
                # Brief pause to prevent busy waiting
                time.sleep(0.01)
                
        except KeyboardInterrupt:
            self.logger.info("Shutdown requested via KeyboardInterrupt")
            self.shutdown_requested = True
            
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
            if self.producer:
                self.logger.info("Flushing producer...")
                self.producer.flush(timeout=10)
                self.producer.close()
                
            if self.consumer:
                self.logger.info("Closing consumer...")
                self.consumer.close()
                
        except Exception as e:
            self.logger.error(f"Error during cleanup: {e}")
            
        self.logger.info(f"Final stats: {self.message_count} messages processed, {self.error_count} errors")
        self.logger.info("Replicator shutdown complete")

if __name__ == "__main__":
    direction = sys.argv[1] if len(sys.argv) > 1 else "S2T"
    replicator = KafkaReplicator(direction)
    replicator.run()