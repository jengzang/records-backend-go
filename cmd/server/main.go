package main

import (
	"log"

	"github.com/jengzang/records-backend-go/internal/api"
	"github.com/jengzang/records-backend-go/internal/config"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 初始化路由
	router := api.SetupRouter(cfg)

	// 启动服务器
	log.Printf("Server starting on port %s", cfg.Port)
	if err := router.Run(cfg.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
