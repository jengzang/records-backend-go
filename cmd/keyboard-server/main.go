package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/jengzang/records-backend-go/internal/keyboard"
)

func main() {
	// Initialize Gin router
	r := gin.Default()

	// Enable CORS
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Database path
	dbPath := "C:/Users/joengzaang/CodeProject/records/keyboard/kmcounter.db"

	// Initialize keyboard handler
	keyboardHandler, err := keyboard.NewHandler(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize keyboard handler: %v", err)
	}
	defer keyboardHandler.Close()

	// Register routes
	api := r.Group("/api")
	keyboardHandler.RegisterRoutes(api)

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Start server
	port := 8080
	fmt.Printf("Server starting on port %d...\n", port)
	if err := r.Run(fmt.Sprintf(":%d", port)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
