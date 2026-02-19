package config

import (
	"os"
)

// Config 应用配置
type Config struct {
	Port       string
	DBPath     string
	JWTSecret  string
	MaxMemory  int64 // 最大内存使用（字节）
}

// Load 加载配置
func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = ":8080"
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data/tracks/tracks.db"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-secret-key-change-in-production"
	}

	return &Config{
		Port:      port,
		DBPath:    dbPath,
		JWTSecret: jwtSecret,
		MaxMemory: 1024 * 1024 * 800, // 800MB 最大内存使用
	}
}
