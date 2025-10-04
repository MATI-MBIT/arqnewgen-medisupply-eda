package drivingadapters

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/MATI-MBIT/arqnewgen-medisupply-eda/simple-service/oder/src/application"
	"github.com/gin-gonic/gin"
)

// ApiServiceAdapter is responsible for exposing the order management capabilities
// over HTTP protocol through RESTful web service endpoints
type ApiServiceAdapter struct {
	server       *http.Server
	router       *gin.Engine
	port         string
	orderService *application.OrderService
}

// CreateOrderRequest represents the request payload for creating an order
type CreateOrderRequest struct {
	CustomerID  string  `json:"customer_id" binding:"required"`
	ProductID   string  `json:"product_id" binding:"required"`
	Quantity    int     `json:"quantity" binding:"required,min=1"`
	TotalAmount float64 `json:"total_amount" binding:"required,min=0"`
}

// UpdateOrderStatusRequest represents the request payload for updating order status
type UpdateOrderStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// NewApiServiceAdapter creates a new ApiServiceAdapter
func NewApiServiceAdapter(port string, orderService *application.OrderService) *ApiServiceAdapter {
	// Set gin to release mode for production
	gin.SetMode(gin.ReleaseMode)
	
	router := gin.New()
	
	// Add middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	
	adapter := &ApiServiceAdapter{
		router:       router,
		port:         port,
		orderService: orderService,
	}
	
	// Setup routes
	adapter.setupRoutes()
	
	// Create HTTP server
	adapter.server = &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}
	
	return adapter
}

// setupRoutes configures all HTTP routes
func (adapter *ApiServiceAdapter) setupRoutes() {
	// Health check endpoint
	adapter.router.GET("/health", adapter.healthHandler)
	
	// Order management endpoints
	v1 := adapter.router.Group("/api/v1")
	{
		v1.POST("/orders", adapter.createOrderHandler)
		v1.GET("/orders", adapter.getAllOrdersHandler)
		v1.GET("/orders/:id", adapter.getOrderHandler)
		v1.PUT("/orders/:id/status", adapter.updateOrderStatusHandler)
	}
}

// healthHandler handles health check requests
func (adapter *ApiServiceAdapter) healthHandler(c *gin.Context) {
	response := gin.H{
		"status":    "healthy",
		"service":   "oder-management/order",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	
	c.JSON(http.StatusOK, response)
}

// createOrderHandler handles order creation requests
func (adapter *ApiServiceAdapter) createOrderHandler(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order, err := adapter.orderService.CreateOrder(req.CustomerID, req.ProductID, req.Quantity, req.TotalAmount)
	if err != nil {
		log.Printf("Error creating order: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
		return
	}

	c.JSON(http.StatusCreated, order)
}

// getAllOrdersHandler handles requests to get all orders
func (adapter *ApiServiceAdapter) getAllOrdersHandler(c *gin.Context) {
	orders, err := adapter.orderService.GetAllOrders()
	if err != nil {
		log.Printf("Error getting orders: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get orders"})
		return
	}

	c.JSON(http.StatusOK, orders)
}

// getOrderHandler handles requests to get a specific order
func (adapter *ApiServiceAdapter) getOrderHandler(c *gin.Context) {
	id := c.Param("id")
	
	order, err := adapter.orderService.GetOrder(id)
	if err != nil {
		log.Printf("Error getting order %s: %v", id, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	c.JSON(http.StatusOK, order)
}

// updateOrderStatusHandler handles requests to update order status
func (adapter *ApiServiceAdapter) updateOrderStatusHandler(c *gin.Context) {
	id := c.Param("id")
	
	var req UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order, err := adapter.orderService.UpdateOrderStatus(id, req.Status)
	if err != nil {
		log.Printf("Error updating order status %s: %v", id, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	c.JSON(http.StatusOK, order)
}

// Start begins the HTTP server
func (adapter *ApiServiceAdapter) Start(ctx context.Context) {
	log.Printf("Starting HTTP API service adapter on port %s...", adapter.port)
	
	// Start server in a goroutine
	go func() {
		if err := adapter.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()
	
	// Wait for context cancellation
	<-ctx.Done()
	log.Println("HTTP API service adapter stopping...")
	
	// Graceful shutdown
	adapter.Stop()
}

// Stop gracefully shuts down the HTTP server
func (adapter *ApiServiceAdapter) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := adapter.server.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	} else {
		log.Println("HTTP API service adapter stopped gracefully")
	}
}