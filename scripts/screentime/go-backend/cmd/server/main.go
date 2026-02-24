package main

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/jengzang/records-backend-go/internal/screentime"
)

func main() {
	// Initialize device manager
	dataDir := "./data/screentime"
	devicesDBPath := filepath.Join(dataDir, "devices.db")

	deviceManager, err := screentime.NewDeviceManager(devicesDBPath, dataDir)
	if err != nil {
		log.Fatalf("Failed to initialize device manager: %v", err)
	}
	defer deviceManager.Close()

	// Create router
	router := gin.Default()

	// Enable CORS
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Create multi-device handler
	handler := screentime.NewMultiDeviceHandler(deviceManager)

	// API routes
	api := router.Group("/api/v1")
	{
		screentimeGroup := api.Group("/screentime")
		{
			// Device management
			screentimeGroup.GET("/devices", handler.ListDevices)

			// Statistics endpoints (support device parameter)
			screentimeGroup.GET("/summary", handler.GetSummary)
			screentimeGroup.GET("/daily", handler.GetDailyStats)
			screentimeGroup.GET("/rankings", handler.GetRankings)

			// TODO: Add more endpoints
			// screentimeGroup.GET("/categories", handler.GetCategories)
			// screentimeGroup.GET("/hourly", handler.GetHourlyStats)
			// screentimeGroup.GET("/trends", handler.GetTrends)
			// screentimeGroup.GET("/app/:packageId", handler.GetAppDetail)
		}
	}

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"service": "records-backend-go",
		})
	})

	// Start server
	port := 8080
	fmt.Printf("Starting server on port %d...\n", port)
	fmt.Printf("API endpoints:\n")
	fmt.Printf("  GET /api/v1/screentime/devices\n")
	fmt.Printf("  GET /api/v1/screentime/summary?device=phone_vivo_x90|computer_main|all\n")
	fmt.Printf("  GET /api/v1/screentime/daily?device=phone_vivo_x90&start=20240101&end=20241231\n")
	fmt.Printf("  GET /api/v1/screentime/rankings?device=phone_vivo_x90&limit=20\n")

	if err := router.Run(fmt.Sprintf(":%d", port)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
