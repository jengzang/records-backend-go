package importer

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"
)

// ImportResult contains the result of an import operation
type ImportResult struct {
	TotalRecords     int
	NewRecords       int
	DuplicateRecords int
}

// CSVImporter handles CSV file imports
type CSVImporter struct {
	db *sql.DB
}

// NewCSVImporter creates a new CSV importer
func NewCSVImporter(db *sql.DB) *CSVImporter {
	return &CSVImporter{db: db}
}

// ImportCSV imports a CSV file into the database
func (i *CSVImporter) ImportCSV(filePath string, mode string, deduplicate bool) (*ImportResult, error) {
	// Open CSV file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Read header
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	// Validate header
	expectedHeaders := []string{"dataTime", "longitude", "latitude", "heading", "accuracy", "speed", "distance", "altitude"}
	if !validateHeader(header, expectedHeaders) {
		return nil, fmt.Errorf("invalid CSV header, expected: %v", expectedHeaders)
	}

	// Start transaction
	tx, err := i.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Clear table if replace mode
	if mode == "replace" {
		if _, err := tx.Exec("DELETE FROM \"一生足迹\""); err != nil {
			return nil, fmt.Errorf("failed to clear table: %w", err)
		}
	}

	result := &ImportResult{}

	// Prepare insert statement
	insertStmt, err := tx.Prepare(`
		INSERT INTO "一生足迹" (dataTime, longitude, latitude, heading, accuracy, speed, distance, altitude, time_visually, time)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare insert statement: %w", err)
	}
	defer insertStmt.Close()

	// Check for duplicates if deduplicate is enabled
	var checkStmt *sql.Stmt
	if deduplicate {
		checkStmt, err = tx.Prepare(`
			SELECT COUNT(*) FROM "一生足迹" WHERE dataTime = ? AND longitude = ? AND latitude = ?
		`)
		if err != nil {
			return nil, fmt.Errorf("failed to prepare check statement: %w", err)
		}
		defer checkStmt.Close()
	}

	// Read and insert records
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read record: %w", err)
		}

		result.TotalRecords++

		// Parse record
		dataTime, err := strconv.ParseInt(record[0], 10, 64)
		if err != nil {
			continue // Skip invalid records
		}

		longitude, _ := strconv.ParseFloat(record[1], 64)
		latitude, _ := strconv.ParseFloat(record[2], 64)
		heading, _ := strconv.ParseFloat(record[3], 64)
		accuracy, _ := strconv.ParseFloat(record[4], 64)
		speed, _ := strconv.ParseFloat(record[5], 64)
		distance, _ := strconv.ParseFloat(record[6], 64)
		altitude, _ := strconv.ParseFloat(record[7], 64)

		// Check for duplicates
		if deduplicate && checkStmt != nil {
			var count int
			err = checkStmt.QueryRow(dataTime, longitude, latitude).Scan(&count)
			if err != nil {
				return nil, fmt.Errorf("failed to check duplicate: %w", err)
			}
			if count > 0 {
				result.DuplicateRecords++
				continue
			}
		}

		// Convert timestamp to formatted strings
		t := time.Unix(dataTime, 0)
		timeVisually := t.Format("2006/01/02 15:04:05.000")
		timeCompact := t.Format("20060102150405")

		// Insert record
		_, err = insertStmt.Exec(dataTime, longitude, latitude, heading, accuracy, speed, distance, altitude, timeVisually, timeCompact)
		if err != nil {
			return nil, fmt.Errorf("failed to insert record: %w", err)
		}

		result.NewRecords++
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return result, nil
}

// validateHeader checks if the CSV header matches expected columns
func validateHeader(header []string, expected []string) bool {
	if len(header) != len(expected) {
		return false
	}
	for i, h := range header {
		if h != expected[i] {
			return false
		}
	}
	return true
}
