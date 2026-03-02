package main

import (
	"database/sql"
	"log"
	"path/filepath"

	"github.com/jengzang/records-backend-go/internal/database"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Database path
	dbPath := filepath.Join("data", "applehealth", "health.db")

	// Open database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Connected to database:", dbPath)

	// Migrations path
	migrationsPath := filepath.Join("scripts", "applehealth", "migrations")

	// Create migration manager
	manager := database.NewMigrationManager(db, migrationsPath)

	// Run migrations
	log.Println("Running migrations from:", migrationsPath)
	if err := manager.RunMigrations(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	log.Println("✅ All migrations completed successfully!")
}
