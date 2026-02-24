package main

import (
	"log"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/jengzang/records-backend-go/internal/screentime"
)

func main() {
	// Initialize screentime device manager
	dataDir := "./data/screentime"
	devicesDBPath := filepath.Join(dataDir, "devices.db")

	deviceManager, err := screentime.NewDeviceManager(devicesDBPath, dataDir)
	if err != nil {
		log.Printf("Warning: Failed to initialize screentime device manager: %v", err)
		deviceManager = nil
	}
	if deviceManager != nil {
		defer deviceManager.Close()
		log.Println("Screentime device manager initialized successfully")
	}

	// 创建 Gin 路由
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

	// 健康检查端点
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"message": "Records Backend API is running",
		})
	})

	// API 路由组
	api := r.Group("/api/v1")
	{
		// 轨迹相关接口
		api.GET("/tracks", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "tracks endpoint"})
		})

		// 键盘鼠标统计接口
		api.GET("/keyboard", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "keyboard endpoint"})
		})

		// 飞机火车路线接口
		api.GET("/flights", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "flights endpoint"})
		})

		// 屏幕使用时间接口 (Multi-device support)
		if deviceManager != nil {
			screentimeHandler := screentime.NewMultiDeviceHandler(deviceManager)
			screentimeGroup := api.Group("/screentime")
			{
				// Device management
				screentimeGroup.GET("/devices", screentimeHandler.ListDevices)

				// Basic statistics
				screentimeGroup.GET("/summary", screentimeHandler.GetSummary)
				screentimeGroup.GET("/daily", screentimeHandler.GetDailyStats)
				screentimeGroup.GET("/rankings", screentimeHandler.GetRankings)

				// Cross-device analysis
				crossDeviceGroup := screentimeGroup.Group("/cross-device")
				{
					crossDeviceGroup.GET("/comparison", screentimeHandler.GetCrossDeviceComparison)
					crossDeviceGroup.GET("/work-life-balance", screentimeHandler.GetWorkLifeBalance)
					crossDeviceGroup.GET("/total-screentime", screentimeHandler.GetTotalScreentime)
					crossDeviceGroup.GET("/switching-patterns", screentimeHandler.GetDeviceSwitchingPatterns)
					crossDeviceGroup.GET("/app-ecosystem", screentimeHandler.GetAppEcosystem)
					crossDeviceGroup.GET("/time-allocation", screentimeHandler.GetTimeAllocation)
					crossDeviceGroup.GET("/user-profile", screentimeHandler.GetUserProfile)
					crossDeviceGroup.GET("/productivity-deep", screentimeHandler.GetProductivityDeep)
					crossDeviceGroup.GET("/focus-analysis", screentimeHandler.GetFocusAnalysis)
					crossDeviceGroup.GET("/recommendations", screentimeHandler.GetCrossDeviceRecommendations)
				}
			}
		} else {
			api.GET("/screentime", func(c *gin.Context) {
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "screentime service not available"})
			})
		}

		// Apple健康数据接口
		api.GET("/health-data", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "health data endpoint"})
		})
	}

	// 启动服务器
	port := ":8080"
	log.Printf("Server starting on port %s", port)
	if err := r.Run(port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
