package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

func main() {
	dbPath := "./data/keyboard/kmcounter.db"
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	fmt.Println("Database connection successful!")

	// Try a simple query
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM daily_stats").Scan(&count)
	if err != nil {
		log.Fatal("Failed to query:", err)
	}

	fmt.Printf("Found %d records in daily_stats\n", count)
}
