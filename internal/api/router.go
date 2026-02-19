package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jengzang/records-backend-go/internal/config"
	"github.com/jengzang/records-backend-go/internal/database"
	"github.com/jengzang/records-backend-go/internal/handler"
	"github.com/jengzang/records-backend-go/internal/middleware"
	"github.com/jengzang/records-backend-go/internal/repository"
	"github.com/jengzang/records-backend-go/internal/service"
)

// SetupRouter 设置路由
func SetupRouter(cfg *config.Config) *gin.Engine {
	// Create Gin engine without default middleware
	r := gin.New()

	// Add custom middleware
	r.Use(middleware.Logger())
	r.Use(middleware.CORS())
	r.Use(middleware.RateLimit(3, time.Second)) // 3 requests per second
	r.Use(gin.Recovery())

	// Initialize database
	db := database.GetDB()

	// Initialize repositories
	trackRepo := repository.NewTrackRepository(db)
	statsRepo := repository.NewStatsRepository(db)
	geocodingRepo := repository.NewGeocodingRepository(db)
	analysisTaskRepo := repository.NewAnalysisTaskRepository(db)

	// Initialize services
	trackService := service.NewTrackService(trackRepo)
	statsService := service.NewStatsService(statsRepo)
	geocodingService := service.NewGeocodingService(geocodingRepo)
	analysisTaskService := service.NewAnalysisTaskService(analysisTaskRepo)

	// Initialize handlers
	trackHandler := handler.NewTrackHandler(trackService)
	statsHandler := handler.NewStatsHandler(statsService)
	geocodingHandler := handler.NewGeocodingHandler(geocodingService)
	analysisTaskHandler := handler.NewAnalysisTaskHandler(analysisTaskService)

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "Records Backend API is running",
		})
	})

	// API 路由组
	api := r.Group("/api/v1")
	{
		// 轨迹相关接口
		tracks := api.Group("/tracks")
		{
			// Track points endpoints
			tracks.GET("/points", trackHandler.GetTrackPoints)
			tracks.GET("/points/:id", trackHandler.GetTrackPointByID)
			tracks.GET("/ungeocoded", trackHandler.GetUngeocodedPoints)

			// Statistics endpoints
			stats := tracks.Group("/statistics")
			{
				stats.GET("/footprint", statsHandler.GetFootprintStatistics)
				stats.GET("/time-distribution", statsHandler.GetTimeDistribution)
				stats.GET("/speed-distribution", statsHandler.GetSpeedDistribution)
			}
		}

		// 键盘鼠标统计接口 (placeholder)
		keyboard := api.Group("/keyboard")
		{
			keyboard.GET("/stats", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "keyboard stats - not implemented yet"})
			})
		}

		// 飞机火车路线接口 (placeholder)
		flights := api.Group("/flights")
		{
			flights.GET("", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "flights list - not implemented yet"})
			})
		}

		// 屏幕使用时间接口 (placeholder)
		screentime := api.Group("/screentime")
		{
			screentime.GET("/stats", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "screentime stats - not implemented yet"})
			})
		}

		// Apple健康数据接口 (placeholder)
		healthData := api.Group("/health-data")
		{
			healthData.GET("/stats", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "health data stats - not implemented yet"})
			})
		}

		// 管理员接口
		admin := api.Group("/admin")
		{
			// Geocoding tasks management
			geocoding := admin.Group("/geocoding")
			{
				geocoding.POST("/tasks", geocodingHandler.CreateTask)
				geocoding.GET("/tasks", geocodingHandler.ListTasks)
				geocoding.GET("/tasks/:id", geocodingHandler.GetTask)
				geocoding.DELETE("/tasks/:id", geocodingHandler.CancelTask)
			}

			// Analysis tasks management
			analysis := admin.Group("/analysis")
			{
				analysis.POST("/tasks", analysisTaskHandler.CreateTask)
				analysis.GET("/tasks", analysisTaskHandler.ListTasks)
				analysis.GET("/tasks/:id", analysisTaskHandler.GetTask)
				analysis.DELETE("/tasks/:id", analysisTaskHandler.CancelTask)
				analysis.POST("/trigger-chain", analysisTaskHandler.TriggerAnalysisChain)
			}
		}
	}

	return r
}
