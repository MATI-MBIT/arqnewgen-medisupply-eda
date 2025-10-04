package drivingadapters

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ApiServiceAdapter is responsible for exposing the application's capabilities
// over HTTP protocol through RESTful web service endpoints
type ApiServiceAdapter struct {
	server *http.Server
	router *gin.Engine
	port   string
}

// NewApiServiceAdapter creates a new ApiServiceAdapter
func NewApiServiceAdapter(port string) *ApiServiceAdapter {
	// Set gin to release mode for production
	gin.SetMode(gin.ReleaseMode)
	
	router := gin.New()
	
	// Add middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	
	adapter := &ApiServiceAdapter{
		router: router,
		port:   port,
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