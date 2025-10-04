# Warehouse Batch Service - Test Scripts

## test_local.py

A comprehensive Python script for building, testing, and running the warehouse batch service locally.

### Usage

```bash
python3 scripts/test_local.py [command] [options]
```

### Commands

- `all` (default) - Run full test suite (deps, validate, build, test)
- `build` - Build the Go application
- `test` - Run Go unit tests
- `run` - Build and run the service
- `clean` - Clean build artifacts
- `deps` - Check dependencies (Go, go.mod)
- `config` - Show current configuration
- `validate` - Validate example JSON files

### Options

- `--env` - Load environment variables from .env file

### Examples

```bash
# Run full test suite
python3 scripts/test_local.py

# Build only
python3 scripts/test_local.py build

# Run tests only
python3 scripts/test_local.py test

# Show configuration
python3 scripts/test_local.py config

# Load .env and run full suite
python3 scripts/test_local.py all --env

# Build and run the service
python3 scripts/test_local.py run
```

### Features

- ✅ Cross-platform Python script
- 🔨 Go application building
- 🧪 Unit test execution
- 📋 JSON validation for example files
- ⚙️ Environment configuration management
- 🧹 Build artifact cleanup
- 🔍 Dependency checking
- 🚀 Service execution

## Additional Scripts

### send_test_event.py

Script to send test order events to Kafka for testing the warehouse batch service.

**Installation:**
```bash
pip install -r scripts/requirements.txt
```

**Usage:**
```bash
# Send a single damage event
python3 scripts/send_test_event.py

# Send a creation event
python3 scripts/send_test_event.py --event-type order.created

# Send multiple events
python3 scripts/send_test_event.py --count 5

# Send to different broker/topic
python3 scripts/send_test_event.py --broker kafka:9092 --topic order-events
```

### Environment Variables

The script sets these default environment variables:

- `KAFKA_ORDER_EVENTS_TOPIC=order-events`
- `KAFKA_BROKER_ADDRESS=localhost:9092`
- `KAFKA_GROUP_ID=warehouse-batch-service`
- `HTTP_PORT=8080`

Use `--env` flag to load from `.env` file instead.