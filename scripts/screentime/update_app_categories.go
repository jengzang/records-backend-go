package main

import (
	"database/sql"
	"fmt"
	"log"
	"path/filepath"

	"github.com/jengzang/records-backend-go/internal/screentime"
	_ "modernc.org/sqlite"
)

func main() {
	mapper := screentime.NewAppCategoryMapper()

	// Get the data directory path
	dataDir := filepath.Join("..", "..", "data", "screentime")

	// Update phone apps
	phoneDBPath := filepath.Join(dataDir, "phone_vivo_x90.db")
	log.Printf("Updating phone app categories at: %s\n", phoneDBPath)

	phoneDB, err := sql.Open("sqlite", phoneDBPath)
	if err != nil {
		log.Fatalf("Failed to open phone database: %v", err)
	}
	defer phoneDB.Close()

	log.Println("Updating phone app categories...")
	if err := mapper.UpdateAppCategories(phoneDB, "screentime_apps", "app_name"); err != nil {
		log.Fatalf("Failed to update phone categories: %v", err)
	}

	// Verify phone updates
	var phoneCount int
	phoneDB.QueryRow("SELECT COUNT(*) FROM screentime_apps WHERE category IS NOT NULL AND category != 'Other'").Scan(&phoneCount)
	log.Printf("Phone: %d apps categorized\n", phoneCount)

	// Update computer apps
	computerDBPath := filepath.Join(dataDir, "manictime_computer.db")
	log.Printf("Updating computer app categories at: %s\n", computerDBPath)

	computerDB, err := sql.Open("sqlite", computerDBPath)
	if err != nil {
		log.Fatalf("Failed to open computer database: %v", err)
	}
	defer computerDB.Close()

	log.Println("Updating computer app categories...")
	if err := mapper.UpdateAppCategories(computerDB, "manictime_apps", "application"); err != nil {
		log.Fatalf("Failed to update computer categories: %v", err)
	}

	// Verify computer updates
	var computerCount int
	computerDB.QueryRow("SELECT COUNT(*) FROM manictime_apps WHERE category IS NOT NULL AND category != 'Other'").Scan(&computerCount)
	log.Printf("Computer: %d apps categorized\n", computerCount)

	// Show category distribution for phone
	log.Println("\n=== Phone Category Distribution ===")
	showCategoryDistribution(phoneDB, "screentime_apps")

	// Show category distribution for computer
	log.Println("\n=== Computer Category Distribution ===")
	showCategoryDistribution(computerDB, "manictime_apps")

	log.Println("\n✅ Category update completed successfully!")
}

func showCategoryDistribution(db *sql.DB, tableName string) {
	query := fmt.Sprintf("SELECT category, COUNT(*) as count FROM %s GROUP BY category ORDER BY count DESC", tableName)
	rows, err := db.Query(query)
	if err != nil {
		log.Printf("Failed to query category distribution: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var category string
		var count int
		if err := rows.Scan(&category, &count); err != nil {
			continue
		}
		log.Printf("  %s: %d apps", category, count)
	}
}
