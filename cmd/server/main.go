package main

import (
	"log"

	"github.com/jengzang/records-backend-go/internal/api"
	"github.com/jengzang/records-backend-go/internal/config"
	"github.com/jengzang/records-backend-go/internal/database"

	// Import analyzer packages to register them
	_ "github.com/jengzang/records-backend-go/internal/analysis/annotation"
	_ "github.com/jengzang/records-backend-go/internal/analysis/behavior"
	_ "github.com/jengzang/records-backend-go/internal/analysis/foundation"
	_ "github.com/jengzang/records-backend-go/internal/analysis/spatial"
	_ "github.com/jengzang/records-backend-go/internal/analysis/stats"
	_ "github.com/jengzang/records-backend-go/internal/analysis/temporal"
	_ "github.com/jengzang/records-backend-go/internal/analysis/viz"
	_ "github.com/jengzang/records-backend-go/internal/analysis/python" // Phase 5 Python workers
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
