package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

func main() {
	db, err := sql.Open("sqlite", "data/keyboard/kmcounter.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT DISTINCT s.scan_code
		FROM scan_codes s
		LEFT JOIN scancode_mapping m ON s.scan_code = m.scancode
		WHERE m.scancode IS NULL
		ORDER BY s.scan_code
		LIMIT 50
	`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("Unmapped scancodes:")
	for rows.Next() {
		var code int
		if err := rows.Scan(&code); err != nil {
			log.Fatal(err)
		}
		fmt.Println(code)
	}
}
