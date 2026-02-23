package importer

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/xuri/excelize/v2"
)

// ExcelImporter handles Excel file imports
type ExcelImporter struct {
	db *sql.DB
}

// NewExcelImporter creates a new Excel importer
func NewExcelImporter(db *sql.DB) *ExcelImporter {
	return &ExcelImporter{db: db}
}

// ImportExcel imports an Excel file into the database
func (e *ExcelImporter) ImportExcel(filePath string, mode string, deduplicate bool) (*ImportResult, error) {
	// Open Excel file
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer f.Close()

	// Get the first sheet
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("no sheets found in Excel file")
	}

	rows, err := f.GetRows(sheets[0])
	if err != nil {
		return nil, fmt.Errorf("failed to read rows: %w", err)
	}

	if len(rows) < 2 {
		return nil, fmt.Errorf("Excel file must have at least header and one data row")
	}

	// Validate header
	header := rows[0]
	expectedHeaders := []string{"dataTime", "longitude", "latitude", "heading", "accuracy", "speed", "distance", "altitude"}
	if !validateHeader(header, expectedHeaders) {
		return nil, fmt.Errorf("invalid Excel header, expected: %v", expectedHeaders)
	}

	// Start transaction
	tx, err := e.db.Begin()
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

	// Process data rows (skip header)
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		if len(row) < 8 {
			continue // Skip incomplete rows
		}

		result.TotalRecords++

		// Parse record
		dataTime, err := strconv.ParseInt(row[0], 10, 64)
		if err != nil {
			continue // Skip invalid records
		}

		longitude, _ := strconv.ParseFloat(row[1], 64)
		latitude, _ := strconv.ParseFloat(row[2], 64)
		heading, _ := strconv.ParseFloat(row[3], 64)
		accuracy, _ := strconv.ParseFloat(row[4], 64)
		speed, _ := strconv.ParseFloat(row[5], 64)
		distance, _ := strconv.ParseFloat(row[6], 64)
		altitude, _ := strconv.ParseFloat(row[7], 64)

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
