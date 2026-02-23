package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "modernc.org/sqlite"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatal("Usage: migrate <db_path> <sql_file>")
	}

	dbPath := os.Args[1]
	sqlFile := os.Args[2]

	// Read SQL file
	sqlBytes, err := os.ReadFile(sqlFile)
	if err != nil {
		log.Fatalf("Failed to read SQL file: %v", err)
	}

	// Open database
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Execute SQL
	_, err = db.Exec(string(sqlBytes))
	if err != nil {
		log.Fatalf("Failed to execute SQL: %v", err)
	}

	fmt.Println("Migration applied successfully")
}
