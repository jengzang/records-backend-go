package main

import (
	"context"
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"

	// Import analyzer packages to register them
	_ "github.com/jengzang/records-backend-go/internal/analysis/stats"
	"github.com/jengzang/records-backend-go/internal/analysis"
)

func main() {
	// Open database
	db, err := sql.Open("sqlite3", "./data/tracks/tracks.db")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	// Test extreme_events analyzer
	log.Println("Testing extreme_events analyzer...")
	analyzer := analysis.GetAnalyzer("extreme_events", db)
	if analyzer == nil {
		log.Fatal("Failed to get extreme_events analyzer")
	}

	ctx := context.Background()
	err = analyzer.Analyze(ctx, 54, "full")
	if err != nil {
		log.Printf("extreme_events analysis failed: %v", err)
	} else {
		log.Println("extreme_events analysis completed successfully")
	}

	// Test admin_crossings analyzer
	log.Println("\nTesting admin_crossings analyzer...")
	analyzer = analysis.GetAnalyzer("admin_crossings", db)
	if analyzer == nil {
		log.Fatal("Failed to get admin_crossings analyzer")
	}

	err = analyzer.Analyze(ctx, 55, "full")
	if err != nil {
		log.Printf("admin_crossings analysis failed: %v", err)
	} else {
		log.Println("admin_crossings analysis completed successfully")
	}

	// Test admin_view_engine analyzer
	log.Println("\nTesting admin_view_engine analyzer...")
	analyzer = analysis.GetAnalyzer("admin_view_engine", db)
	if analyzer == nil {
		log.Fatal("Failed to get admin_view_engine analyzer")
	}

	err = analyzer.Analyze(ctx, 56, "full")
	if err != nil {
		log.Printf("admin_view_engine analysis failed: %v", err)
	} else {
		log.Println("admin_view_engine analysis completed successfully")
	}

	log.Println("\nAll tests completed!")
}
