package main

import (
	"database/sql"
	"fmt"
	"log"
	_ "modernc.org/sqlite"
)

func main() {
	// Check devices.db
	fmt.Println("=== Checking devices.db ===")
	devicesDB, err := sql.Open("sqlite", "./data/screentime/devices.db")
	if err != nil {
		log.Fatal(err)
	}
	defer devicesDB.Close()

	rows, _ := devicesDB.Query("SELECT id, name, type, db_path, data_format, is_active FROM devices")
	defer rows.Close()
	
	for rows.Next() {
		var id, name, typ, dbPath, dataFormat string
		var isActive int
		rows.Scan(&id, &name, &typ, &dbPath, &dataFormat, &isActive)
		fmt.Printf("Device: %s (%s) - Type: %s, DB: %s, Active: %d\n", id, name, typ, dbPath, isActive)
	}

	// Check ManicTime database structure
	fmt.Println("\n=== Checking manictime_computer.db structure ===")
	manicDB, err := sql.Open("sqlite", "./data/screentime/manictime_computer.db")
	if err != nil {
		log.Fatal(err)
	}
	defer manicDB.Close()

	// Get table list
	tables, _ := manicDB.Query("SELECT name FROM sqlite_master WHERE type='table' ORDER BY name")
	defer tables.Close()
	
	fmt.Println("Tables:")
	for tables.Next() {
		var tableName string
		tables.Scan(&tableName)
		fmt.Printf("  - %s\n", tableName)
	}

	// Check manictime_daily structure
	fmt.Println("\n=== manictime_daily table structure ===")
	pragma, _ := manicDB.Query("PRAGMA table_info(manictime_daily)")
	defer pragma.Close()
	
	for pragma.Next() {
		var cid int
		var name, typ string
		var notnull, pk int
		var dfltValue sql.NullString
		pragma.Scan(&cid, &name, &typ, &notnull, &dfltValue, &pk)
		fmt.Printf("  %s (%s)\n", name, typ)
	}

	// Check data sample
	fmt.Println("\n=== Sample data from manictime_daily ===")
	sample, _ := manicDB.Query("SELECT date, application, category, total_duration_seconds FROM manictime_daily LIMIT 5")
	defer sample.Close()
	
	for sample.Next() {
		var date, app, category string
		var duration int
		sample.Scan(&date, &app, &category, &duration)
		fmt.Printf("  %s | %s | %s | %ds\n", date, app, category, duration)
	}

	// Check total records
	var count int
	manicDB.QueryRow("SELECT COUNT(*) FROM manictime_daily").Scan(&count)
	fmt.Printf("\nTotal records in manictime_daily: %d\n", count)
}
