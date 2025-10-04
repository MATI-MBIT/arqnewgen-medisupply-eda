# Batch Implementation for Warehouse Service

## Overview

The warehouse batch service has been updated to handle order events from the order management microservice using a batch domain aggregate pattern. This implementation provides efficient batch processing of orders while maintaining data consistency and traceability.

## Architecture

### Domain Layer

#### Batch Aggregate (`domain/batch.go`)
- **Batch**: Main aggregate root representing a collection of orders for the same product
- **BatchItem**: Value object representing individual orders within a batch
- **BatchStatus**: Enum for batch lifecycle states (pending, processing, completed, cancelled, damaged)

Key features:
- Automatic batch creation for new products
- Order grouping by product ID
- Status tracking for individual orders and entire batches
- Business rule enforcement (e.g., can't add to completed batches)

#### Repository Interface (`domain/batch_repository.go`)
- Defines contracts for batch persistence operations
- Supports queries by ID, product, status, and order

### Application Layer

#### BatchService (`application/batch_service.go`)
- Core business logic for batch operations
- Handles batch lifecycle management
- Provides order-to-batch mapping functionality

#### OrderService (`application/order_service.go`)
- Updated to integrate with BatchService
- Processes order events and updates corresponding batches
- Handles various order states (allocated, shipped, delivered, damaged, etc.)

### Infrastructure Layer

#### In-Memory Repository (`infrastructure/driven-adapters/batch_memory_repository.go`)
- Thread-safe in-memory implementation of BatchRepository
- Provides deep copying to prevent external modifications
- Suitable for development and testing

#### API Endpoints (`infrastructure/driving-adapters/api_service_adapter.go`)
- RESTful endpoints for batch management
- Supports querying batches by various criteria

## API Endpoints

### Batch Management
- `GET /api/v1/batches` - Get all batches
- `GET /api/v1/batches/product/:productId` - Get batches for a specific product
- `GET /api/v1/batches/status/:status` - Get batches by status
- `GET /api/v1/batches/order/:orderId` - Get batch containing a specific order

### Health Check
- `GET /health` - Service health status

## Order Event Processing

The service processes the following order events:

| Event Type | Warehouse Action | Batch Operation |
|------------|------------------|-----------------|
| `order.created` | `allocate_inventory` | Add order to batch |
| `order.cancelled` | `release_inventory` | Remove order from batch |
| `order.shipped` | `update_inventory` | Update order status to shipped |
| `order.delivered` | `confirm_delivery` | Update order status to delivered |
| `order.returned` | `process_return` | Update status + add return entry |
| `order.damage_processed` | `process_damage` | Handle damage scenarios (auto-creates batch if needed) |
| `order.inventory_allocated` | `confirm_allocation` | Confirm allocation |
| `order.inventory_released` | `confirm_release` | Confirm release |

### Damage Processing Special Behavior

The `process_damage` action has special logic to handle orders that may not exist in any batch yet:

- **If order exists in batch**: Updates the order status within the existing batch
- **If order doesn't exist**: Creates a new batch with the order and appropriate damage status
- **Major damage**: Automatically marks the entire batch as damaged
- **Minor damage**: Marks order for inspection but keeps batch in pending status
- **Damage processed**: Indicates completion of damage handling process

## Batch Lifecycle

1. **Pending**: New batch created when first order for a product arrives
2. **Processing**: Batch is being actively processed in warehouse
3. **Completed**: All orders in batch have been processed
4. **Cancelled**: Batch cancelled (e.g., all orders cancelled)
5. **Damaged**: Batch marked as damaged due to product issues

## Key Features

### Automatic Batch Management
- Orders for the same product are automatically grouped into pending batches
- New batches are created when no pending batch exists for a product
- Empty batches are automatically cleaned up when last order is removed

### Smart Damage Processing
- **Auto-batch Creation**: Damage events for non-existing orders automatically create new batches
- **Damage Severity Handling**: Different actions based on damage severity (minor vs major)
- **Batch Status Management**: Major damage automatically marks entire batch as damaged
- **Flexible Processing**: Handles both existing orders and new damage reports seamlessly

### Order Tracking
- Each order maintains its individual status within the batch
- Timestamps for when orders are added and processed
- Full traceability of order lifecycle within warehouse operations

### Concurrent Safety
- Thread-safe repository implementation
- Proper locking mechanisms for concurrent access
- Deep copying to prevent data races

## Testing

The implementation includes comprehensive tests:
- Unit tests for batch domain logic
- Integration tests for batch service operations
- Order event processing tests

Run tests with:
```bash
go test ./...
```

## Configuration

The service uses the existing configuration system with environment variables:
- Kafka broker settings for order event consumption
- HTTP port for API endpoints
- Service-specific settings in `.env` file

## Future Enhancements

Potential improvements for production use:
1. **Persistent Storage**: Replace in-memory repository with database implementation
2. **Batch Size Limits**: Implement maximum batch size constraints
3. **Batch Scheduling**: Automatic batch processing based on time or size triggers
4. **Metrics**: Add monitoring and metrics for batch operations
5. **Event Publishing**: Publish batch events for downstream services
6. **Batch Optimization**: Intelligent batching algorithms based on warehouse capacity

## Dependencies

- Go 1.24.1+
- Gin web framework for HTTP API
- Kafka Go client for event consumption
- Standard library for core functionality