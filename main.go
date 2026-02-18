package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	// 创建 Gin 路由
	r := gin.Default()

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

		// 屏幕使用时间接口
		api.GET("/screentime", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "screentime endpoint"})
		})

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
