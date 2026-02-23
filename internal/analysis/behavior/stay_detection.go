package behavior

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"

	"github.com/jengzang/records-backend-go/internal/analysis"
)

// StayDetectionAnalyzer implements SPATIAL stay detection using DBSCAN
// Skill: 停留检测 (Stay Detection) - SPATIAL type
// Detects stays where GPS points cluster spatially (home, office, GPS drift)
type StayDetectionAnalyzer struct {
	*analysis.IncrementalAnalyzer
}

// NewStayDetectionAnalyzer creates a new stay detection analyzer
func NewStayDetectionAnalyzer(db *sql.DB) analysis.Analyzer {
	return &StayDetectionAnalyzer{
		IncrementalAnalyzer: analysis.NewIncrementalAnalyzer(db, "stay_detection", 1000),
	}
}

// Analyze performs SPATIAL stay detection by calling Python worker
func (a *StayDetectionAnalyzer) Analyze(ctx context.Context, taskID int64, mode string) error {
	log.Printf("[StayDetectionAnalyzer] Starting SPATIAL stay detection (task_id=%d, mode=%s)", taskID, mode)

	// Get database path from connection
	dbPath, err := a.GetDBPath()
	if err != nil {
		return fmt.Errorf("failed to get database path: %w", err)
	}

	// Construct path to Python worker
	workerPath := filepath.Join("scripts", "tracks", "workers", "stay_detection.py")

	// Call Python worker
	cmd := exec.CommandContext(ctx, "python", workerPath, dbPath, fmt.Sprintf("%d", taskID))
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("[StayDetectionAnalyzer] Python worker failed: %s", string(output))
		return fmt.Errorf("stay detection failed: %w, output: %s", err, string(output))
	}

	log.Printf("[StayDetectionAnalyzer] Python worker output: %s", string(output))
	log.Printf("[StayDetectionAnalyzer] SPATIAL stay detection completed")
	return nil
}

// GetDBPath extracts the database file path from the connection
func (a *StayDetectionAnalyzer) GetDBPath() (string, error) {
	// Query for database file path
	var dbPath string
	err := a.DB.QueryRow("PRAGMA database_list").Scan(nil, nil, &dbPath)
	if err != nil {
		// Fallback: use default path
		return "data/tracks.db", nil
	}
	return dbPath, nil
}

// Register the analyzer
func init() {
	analysis.RegisterAnalyzer("stay_detection", NewStayDetectionAnalyzer)
}
