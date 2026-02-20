package python

import (
	"context"
	"database/sql"
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/jengzang/records-backend-go/internal/analysis"
	"github.com/jengzang/records-backend-go/internal/database"
)

// PythonWorkerAnalyzer wraps a Python worker script as an Analyzer
type PythonWorkerAnalyzer struct {
	*analysis.BaseAnalyzer
	scriptPath string
	dbPath     string
}

// NewPythonWorkerAnalyzer creates a new Python worker analyzer
func NewPythonWorkerAnalyzer(db *sql.DB, name string, scriptPath string, dbPath string) *PythonWorkerAnalyzer {
	return &PythonWorkerAnalyzer{
		BaseAnalyzer: analysis.NewBaseAnalyzer(db, name),
		scriptPath:   scriptPath,
		dbPath:       dbPath,
	}
}

// Analyze executes the Python worker script
func (a *PythonWorkerAnalyzer) Analyze(ctx context.Context, taskID int64, mode string) error {
	// Mark task as running
	if err := a.MarkTaskAsRunning(taskID); err != nil {
		return fmt.Errorf("failed to mark task as running: %w", err)
	}

	// Execute Python worker
	cmd := exec.CommandContext(ctx, "python", a.scriptPath, a.dbPath, fmt.Sprintf("%d", taskID))

	// Capture output
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Mark task as failed
		failQuery := `
			UPDATE analysis_tasks
			SET status = 'failed',
			    error_message = ?,
			    completed_at = CURRENT_TIMESTAMP
			WHERE id = ?
		`
		a.DB.Exec(failQuery, string(output), taskID)
		return fmt.Errorf("python worker failed: %w, output: %s", err, string(output))
	}

	// Task completion is handled by the Python worker itself
	return nil
}

// GetProgress returns the current progress from the database
func (a *PythonWorkerAnalyzer) GetProgress(taskID int64) (*analysis.Progress, error) {
	query := `
		SELECT processed_points, total_points, failed_points, progress_percent
		FROM analysis_tasks
		WHERE id = ?
	`

	var processed, total, failed int
	var percent float64

	err := a.DB.QueryRow(query, taskID).Scan(&processed, &total, &failed, &percent)
	if err != nil {
		return nil, err
	}

	return &analysis.Progress{
		Processed: processed,
		Total:     total,
		Failed:    failed,
		Percent:   percent,
	}, nil
}

// init registers all Phase 5 Python workers
func init() {
	// Get database path from config
	dbPath := "./data/tracks/tracks.db" // Default path

	workers := []struct {
		name       string
		scriptName string
	}{
		{"stay_detection", "stay_detection.py"},
		{"density_structure_advanced", "density_structure_advanced.py"},
		{"trip_construction_advanced", "trip_construction_advanced.py"},
		{"spatial_persona", "spatial_persona.py"},
		{"admin_view_advanced", "admin_view_advanced.py"},
	}

	for _, w := range workers {
		workerName := w.name
		scriptName := w.scriptName

		// Register factory function
		analysis.RegisterAnalyzer(workerName, func(db *sql.DB) analysis.Analyzer {
			scriptsDir := "./scripts/tracks"
			scriptPath := filepath.Join(scriptsDir, "workers", scriptName)
			return NewPythonWorkerAnalyzer(db, workerName, scriptPath, dbPath)
		})
	}
}

