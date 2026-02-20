package main

import (
	"log"

	"github.com/jengzang/records-backend-go/internal/api"
	"github.com/jengzang/records-backend-go/internal/config"
	"github.com/jengzang/records-backend-go/internal/database"

	// Import analyzer packages to register them
	_ "github.com/jengzang/records-backend-go/internal/analysis/annotation"
	//_ "github.com/jengzang/records-backend-go/internal/analysis/behavior" // Temporarily disabled - type conflicts
	//_ "github.com/jengzang/records-backend-go/internal/analysis/foundation" // Temporarily disabled - missing Point type
	//_ "github.com/jengzang/records-backend-go/internal/analysis/spatial" // Temporarily disabled - type errors
	//_ "github.com/jengzang/records-backend-go/internal/analysis/stats" // Temporarily disabled - missing methods
	//_ "github.com/jengzang/records-backend-go/internal/analysis/temporal" // Temporarily disabled - unused imports
	_ "github.com/jengzang/records-backend-go/internal/analysis/viz"
	//_ "github.com/jengzang/records-backend-go/internal/analysis/python" // Temporarily disabled - unused imports
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 初始化数据库
	dbConfig := database.Config{
		Path: cfg.DBPath,
	}
	if err := database.Init(dbConfig); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer database.Close()

	// 初始化路由
	router := api.SetupRouter(cfg)

	// 启动服务器
	log.Printf("Server starting on port %s", cfg.Port)
	if err := router.Run(cfg.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
