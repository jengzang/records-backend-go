package config

import (
	"os"
)

// Config 应用配置
type Config struct {
	Port             string
	DBPath           string
	KeyboardDBPath   string
	ScreentimeDBPath string
	FlightsDBPath    string
	RailwayDBPath    string
	JWTSecret        string
	MaxMemory        int64 // 最大内存使用（字节）
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

	keyboardDBPath := os.Getenv("KEYBOARD_DB_PATH")
	if keyboardDBPath == "" {
		keyboardDBPath = "./data/keyboard/kmcounter.db"
	}

	screentimeDBPath := os.Getenv("SCREENTIME_DB_PATH")
	if screentimeDBPath == "" {
		screentimeDBPath = "./data/screentime/screentime.db"
	}

	flightsDBPath := os.Getenv("FLIGHTS_DB_PATH")
	if flightsDBPath == "" {
		flightsDBPath = "./data/flights/flights.db"
	}

	railwayDBPath := os.Getenv("RAILWAY_DB_PATH")
	if railwayDBPath == "" {
		railwayDBPath = "./data/railway/railway.db"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-secret-key-change-in-production"
	}

	return &Config{
		Port:             port,
		DBPath:           dbPath,
		KeyboardDBPath:   keyboardDBPath,
		ScreentimeDBPath: screentimeDBPath,
		FlightsDBPath:    flightsDBPath,
		RailwayDBPath:    railwayDBPath,
		JWTSecret:        jwtSecret,
		MaxMemory:        1024 * 1024 * 800, // 800MB 最大内存使用
	}
}
