#!/bin/bash

# Demo script for Order Management Service
# This script demonstrates the hexagonal architecture implementation

echo "=== Order Management Service Demo ==="
echo

# Check if service is running
echo "1. Checking service health..."
curl -s http://localhost:8081/health | jq '.' || echo "Service not running. Please start with: go run src/main.go"
echo

# Create a new order
echo "2. Creating a new order..."
ORDER_RESPONSE=$(curl -s -X POST http://localhost:8081/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": "customer-123",
    "product_id": "product-456",
    "quantity": 2,
    "total_amount": 99.99
  }')

echo "$ORDER_RESPONSE" | jq '.'
ORDER_ID=$(echo "$ORDER_RESPONSE" | jq -r '.id')
echo "Created order with ID: $ORDER_ID"
echo

# Get all orders
echo "3. Getting all orders..."
curl -s http://localhost:8081/api/v1/orders | jq '.'
echo

# Get specific order
echo "4. Getting order by ID..."
curl -s "http://localhost:8081/api/v1/orders/$ORDER_ID" | jq '.'
echo

# Update order status
echo "5. Updating order status..."
curl -s -X PUT "http://localhost:8081/api/v1/orders/$ORDER_ID/status" \
  -H "Content-Type: application/json" \
  -d '{"status": "shipped"}' | jq '.'
echo

# Get updated order
echo "6. Getting updated order..."
curl -s "http://localhost:8081/api/v1/orders/$ORDER_ID" | jq '.'
echo

echo "=== Demo completed ==="
echo "Check the service logs to see the hexagonal architecture in action:"
echo "- HTTP requests handled by ApiServiceAdapter"
echo "- Business logic processed by OrderService"
echo "- Events published via RabbitMQPublisher"
echo "- Data stored via MemoryOrderRepository"