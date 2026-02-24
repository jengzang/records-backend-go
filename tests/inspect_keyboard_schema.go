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

	tables := []string{"keyboard_data", "mouse_data", "scan_codes"}

	for _, table := range tables {
		fmt.Printf("\n=== Table: %s ===\n", table)

		// Get schema
		rows, err := db.Query(fmt.Sprintf("PRAGMA table_info(%s)", table))
		if err != nil {
			log.Printf("Failed to get schema for %s: %v\n", table, err)
			continue
		}

		fmt.Println("Columns:")
		for rows.Next() {
			var cid int
			var name, typ string
			var notnull, pk int
			var dfltValue sql.NullString

			if err := rows.Scan(&cid, &name, &typ, &notnull, &dfltValue, &pk); err != nil {
				log.Fatal(err)
			}
			fmt.Printf("  %s (%s)\n", name, typ)
		}
		rows.Close()

		// Get row count
		var count int
		db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count)
		fmt.Printf("Row count: %d\n", count)
	}
}
