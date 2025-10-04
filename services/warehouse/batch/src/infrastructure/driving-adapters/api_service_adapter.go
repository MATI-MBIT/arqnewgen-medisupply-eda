package drivingadapters

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/MATI-MBIT/arqnewgen-medisupply-eda/simple-service/batch/src/application"
	"github.com/MATI-MBIT/arqnewgen-medisupply-eda/simple-service/batch/src/domain"
)

// ApiServiceAdapter is responsible for exposing the application's capabilities
// over HTTP protocol through RESTful web service endpoints
type ApiServiceAdapter struct {
	server       *http.Server
	router       *gin.Engine
	port         string
	batchService application.BatchServiceInterface
}

// NewApiServiceAdapter creates a new ApiServiceAdapter
func NewApiServiceAdapter(port string, batchService application.BatchServiceInterface) *ApiServiceAdapter {
	// Set gin to release mode for production
	gin.SetMode(gin.ReleaseMode)
	
	router := gin.New()
	
	// Add middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	
	adapter := &ApiServiceAdapter{
		router:       router,
		port:         port,
		batchService: batchService,
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
	
	// Batch endpoints
	v1 := adapter.router.Group("/api/v1")
	{
		v1.GET("/batches", adapter.getAllBatchesHandler)
		v1.GET("/batches/product/:productId", adapter.getBatchesByProductHandler)
		v1.GET("/batches/status/:status", adapter.getBatchesByStatusHandler)
		v1.GET("/batches/order/:orderId", adapter.getBatchByOrderHandler)
	}
}

// healthHandler handles health check requests
func (adapter *ApiServiceAdapter) healthHandler(c *gin.Context) {
	response := gin.H{
		"status":    "healthy",
		"service":   "warehouse-batch",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	
	c.JSON(http.StatusOK, response)
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

// getAllBatchesHandler handles GET /api/v1/batches
func (adapter *ApiServiceAdapter) getAllBatchesHandler(c *gin.Context) {
	batches, err := adapter.batchService.GetAllBatches()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve batches",
			"details": err.Error(),
		})
		return
	}
	
	batchDTOs := application.ToBatchDTOs(batches)
	c.JSON(http.StatusOK, gin.H{
		"batches": batchDTOs,
		"count":   len(batchDTOs),
	})
}

// getBatchesByProductHandler handles GET /api/v1/batches/product/:productId
func (adapter *ApiServiceAdapter) getBatchesByProductHandler(c *gin.Context) {
	productID := c.Param("productId")
	if productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Product ID is required",
		})
		return
	}
	
	batches, err := adapter.batchService.GetBatchesByProductID(productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve batches for product",
			"details": err.Error(),
		})
		return
	}
	
	batchDTOs := application.ToBatchDTOs(batches)
	c.JSON(http.StatusOK, gin.H{
		"product_id": productID,
		"batches":    batchDTOs,
		"count":      len(batchDTOs),
	})
}

// getBatchesByStatusHandler handles GET /api/v1/batches/status/:status
func (adapter *ApiServiceAdapter) getBatchesByStatusHandler(c *gin.Context) {
	statusStr := c.Param("status")
	if statusStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Status is required",
		})
		return
	}
	
	status := domain.BatchStatus(statusStr)
	batches, err := adapter.batchService.GetBatchesByStatus(status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve batches by status",
			"details": err.Error(),
		})
		return
	}
	
	batchDTOs := application.ToBatchDTOs(batches)
	c.JSON(http.StatusOK, gin.H{
		"status":  status,
		"batches": batchDTOs,
		"count":   len(batchDTOs),
	})
}

// getBatchByOrderHandler handles GET /api/v1/batches/order/:orderId
func (adapter *ApiServiceAdapter) getBatchByOrderHandler(c *gin.Context) {
	orderID := c.Param("orderId")
	if orderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Order ID is required",
		})
		return
	}
	
	batch, err := adapter.batchService.GetBatchByOrderID(orderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Batch not found for order",
			"details": err.Error(),
		})
		return
	}
	
	batchDTO := application.ToBatchDTO(batch)
	c.JSON(http.StatusOK, gin.H{
		"order_id": orderID,
		"batch":    batchDTO,
	})
}