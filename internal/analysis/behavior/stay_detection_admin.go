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

// AdminAreaStayAnalyzer implements ADMIN_AREA stay detection using RLE
// Skill: 停留检测 (Stay Detection) - ADMIN_AREA type
// Detects stays where user remains in the same administrative area
type AdminAreaStayAnalyzer struct {
	*analysis.IncrementalAnalyzer
}

// NewAdminAreaStayAnalyzer creates a new admin area stay analyzer
func NewAdminAreaStayAnalyzer(db *sql.DB) analysis.Analyzer {
	return &AdminAreaStayAnalyzer{
		IncrementalAnalyzer: analysis.NewIncrementalAnalyzer(db, "stay_detection_admin", 1000),
	}
}

// Analyze performs ADMIN_AREA stay detection by calling Python worker
func (a *AdminAreaStayAnalyzer) Analyze(ctx context.Context, taskID int64, mode string) error {
	log.Printf("[AdminAreaStayAnalyzer] Starting ADMIN_AREA stay detection (task_id=%d, mode=%s)", taskID, mode)

	// Get database path from connection
	dbPath, err := a.GetDBPath()
	if err != nil {
		return fmt.Errorf("failed to get database path: %w", err)
	}

	// Construct path to Python worker
	workerPath := filepath.Join("scripts", "tracks", "workers", "stay_detection_admin.py")

	// Call Python worker
	cmd := exec.CommandContext(ctx, "python", workerPath, dbPath, fmt.Sprintf("%d", taskID))
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("[AdminAreaStayAnalyzer] Python worker failed: %s", string(output))
		return fmt.Errorf("admin area stay detection failed: %w, output: %s", err, string(output))
	}

	log.Printf("[AdminAreaStayAnalyzer] Python worker output: %s", string(output))
	log.Printf("[AdminAreaStayAnalyzer] ADMIN_AREA stay detection completed")
	return nil
}

// GetDBPath extracts the database file path from the connection
func (a *AdminAreaStayAnalyzer) GetDBPath() (string, error) {
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
	analysis.RegisterAnalyzer("stay_detection_admin", NewAdminAreaStayAnalyzer)
}
