package main

import (
	"os"

	"github.com/jengzang/records-backend-go/internal/api"
	"github.com/jengzang/records-backend-go/internal/config"
	"github.com/jengzang/records-backend-go/internal/database"
	"github.com/jengzang/records-backend-go/internal/logger"
	"github.com/sirupsen/logrus"

	// Import analyzer packages to register them
	_ "github.com/jengzang/records-backend-go/internal/analysis/advanced"
	_ "github.com/jengzang/records-backend-go/internal/analysis/annotation"
	_ "github.com/jengzang/records-backend-go/internal/analysis/behavior"
	_ "github.com/jengzang/records-backend-go/internal/analysis/foundation"
	_ "github.com/jengzang/records-backend-go/internal/analysis/integration"
	_ "github.com/jengzang/records-backend-go/internal/analysis/spatial"
	_ "github.com/jengzang/records-backend-go/internal/analysis/stats"
	_ "github.com/jengzang/records-backend-go/internal/analysis/temporal"
	_ "github.com/jengzang/records-backend-go/internal/analysis/viz"
	_ "github.com/jengzang/records-backend-go/internal/analysis/python"
)

func main() {
	// Initialize logger
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}
	logger.Init(logLevel)

	logger.Info("Starting Records Backend API", logrus.Fields{
		"version": "1.0.0",
		"port":    ":8080",
	})

	// 加载配置
	cfg := config.Load()

	// 初始化数据库
	dbConfig := database.Config{
		Path: cfg.DBPath,
	}
	if err := database.Init(dbConfig); err != nil {
		logger.Fatal("Failed to initialize database", err, nil)
	}
	defer database.Close()

	logger.Info("Database initialized successfully", logrus.Fields{
		"db_path": cfg.DBPath,
	})

	// 初始化路由
	router := api.SetupRouter(cfg)

	// 启动服务器
	logger.Info("Server starting", logrus.Fields{
		"port": cfg.Port,
	})
	if err := router.Run(cfg.Port); err != nil {
		logger.Fatal("Failed to start server", err, nil)
	}
}
