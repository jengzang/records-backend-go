package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jengzang/records-backend-go/internal/config"
)

// SetupRouter 设置路由
func SetupRouter(cfg *config.Config) *gin.Engine {
	r := gin.Default()

	// CORS 中间件
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	// 健康检查
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
		tracks := api.Group("/tracks")
		{
			tracks.GET("", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "tracks list"})
			})
			tracks.POST("", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "create track"})
			})
		}

		// 键盘鼠标统计接口
		keyboard := api.Group("/keyboard")
		{
			keyboard.GET("/stats", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "keyboard stats"})
			})
		}

		// 飞机火车路线接口
		flights := api.Group("/flights")
		{
			flights.GET("", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "flights list"})
			})
		}

		// 屏幕使用时间接口
		screentime := api.Group("/screentime")
		{
			screentime.GET("/stats", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "screentime stats"})
			})
		}

		// Apple健康数据接口
		health := api.Group("/health-data")
		{
			health.GET("/stats", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "health data stats"})
			})
		}
	}

	return r
}
