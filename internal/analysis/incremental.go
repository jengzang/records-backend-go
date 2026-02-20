package analysis

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// IncrementalAnalyzer provides base functionality for incremental analysis
type IncrementalAnalyzer struct {
	*BaseAnalyzer
	BatchSize int // Number of records to process in each batch
}

// NewIncrementalAnalyzer creates a new incremental analyzer
func NewIncrementalAnalyzer(db *sql.DB, name string, batchSize int) *IncrementalAnalyzer {
	if batchSize <= 0 {
		batchSize = 1000 // Default batch size
	}

	return &IncrementalAnalyzer{
		BaseAnalyzer: NewBaseAnalyzer(db, name),
		BatchSize:    batchSize,
	}
}

// ProcessInBatches processes records in batches with progress tracking
// query: SQL query to fetch records (should support LIMIT and OFFSET)
// processFunc: function to process each batch of records
func (a *IncrementalAnalyzer) ProcessInBatches(
	ctx context.Context,
	taskID int64,
	query string,
	args []interface{},
	processFunc func(rows *sql.Rows) error,
) error {
	// Mark task as running
	if err := a.MarkTaskAsRunning(taskID); err != nil {
		return fmt.Errorf("failed to mark task as running: %w", err)
	}

	// Count total records
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM (%s) AS subquery", query)
	var total int
	if err := a.DB.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return fmt.Errorf("failed to count records: %w", err)
	}

	if total == 0 {
		// No records to process, mark as completed
		return a.MarkTaskAsCompleted(taskID)
	}

	// Process in batches
	processed := 0
	failed := 0
	startTime := time.Now()

	for offset := 0; offset < total; offset += a.BatchSize {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Fetch batch
		batchQuery := fmt.Sprintf("%s LIMIT %d OFFSET %d", query, a.BatchSize, offset)
		rows, err := a.DB.QueryContext(ctx, batchQuery, args...)
		if err != nil {
			return fmt.Errorf("failed to fetch batch at offset %d: %w", offset, err)
		}

		// Process batch
		batchFailed := 0
		err = processFunc(rows)
		rows.Close()

		if err != nil {
			// Count this batch as failed but continue
			batchFailed = a.BatchSize
			if offset+a.BatchSize > total {
				batchFailed = total - offset
			}
			failed += batchFailed
		}

		// Update progress
		processed = offset + a.BatchSize
		if processed > total {
			processed = total
		}

		if err := a.UpdateTaskProgress(taskID, int64(processed), int64(total), int64(failed)); err != nil {
			return fmt.Errorf("failed to update progress: %w", err)
		}

		// Calculate ETA
		elapsed := time.Since(startTime).Seconds()
		if processed > 0 {
			eta := int(elapsed / float64(processed) * float64(total-processed))
			// Could store ETA in database if needed
			_ = eta
		}
	}

	// Mark as completed
	if err := a.MarkTaskAsCompleted(taskID); err != nil {
		return fmt.Errorf("failed to mark task as completed: %w", err)
	}

	return nil
}

// ProcessIncrementalBatches processes only new/unanalyzed records
// tableName: name of the table to process
// whereClause: additional WHERE conditions for incremental processing
// processFunc: function to process each batch
func (a *IncrementalAnalyzer) ProcessIncrementalBatches(
	ctx context.Context,
	taskID int64,
	tableName string,
	whereClause string,
	processFunc func(rows *sql.Rows) error,
) error {
	query := fmt.Sprintf("SELECT * FROM %s WHERE %s", tableName, whereClause)
	return a.ProcessInBatches(ctx, taskID, query, nil, processFunc)
}

// GetLastProcessedTimestamp retrieves the timestamp of the last processed record
// for incremental analysis
func (a *IncrementalAnalyzer) GetLastProcessedTimestamp(tableName string) (int64, error) {
	query := fmt.Sprintf(`
		SELECT MAX(updated_at)
		FROM %s
		WHERE analysis_version IS NOT NULL
	`, tableName)

	var timestamp sql.NullInt64
	err := a.DB.QueryRow(query).Scan(&timestamp)
	if err != nil {
		return 0, err
	}

	if !timestamp.Valid {
		return 0, nil
	}

	return timestamp.Int64, nil
}

// MarkRecordsAsAnalyzed marks records as analyzed with the current algorithm version
func (a *IncrementalAnalyzer) MarkRecordsAsAnalyzed(
	tableName string,
	recordIDs []int64,
	algoVersion string,
) error {
	if len(recordIDs) == 0 {
		return nil
	}

	// Build placeholders for IN clause
	placeholders := ""
	args := make([]interface{}, len(recordIDs)+1)
	for i := range recordIDs {
		if i > 0 {
			placeholders += ","
		}
		placeholders += "?"
		args[i] = recordIDs[i]
	}
	args[len(recordIDs)] = algoVersion

	query := fmt.Sprintf(`
		UPDATE %s
		SET analysis_version = ?,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id IN (%s)
	`, tableName, placeholders)

	_, err := a.DB.Exec(query, args...)
	return err
}

// BatchInsert performs a batch insert operation
func (a *IncrementalAnalyzer) BatchInsert(
	tableName string,
	columns []string,
	values [][]interface{},
) error {
	if len(values) == 0 {
		return nil
	}

	// Build INSERT query
	query := fmt.Sprintf("INSERT INTO %s (", tableName)
	for i, col := range columns {
		if i > 0 {
			query += ", "
		}
		query += col
	}
	query += ") VALUES "

	// Build placeholders
	args := []interface{}{}
	for i, row := range values {
		if i > 0 {
			query += ", "
		}
		query += "("
		for j := range row {
			if j > 0 {
				query += ", "
			}
			query += "?"
			args = append(args, row[j])
		}
		query += ")"
	}

	_, err := a.DB.Exec(query, args...)
	return err
}

// BatchUpdate performs a batch update operation
func (a *IncrementalAnalyzer) BatchUpdate(
	tableName string,
	setClause string,
	whereClause string,
	args []interface{},
) error {
	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s", tableName, setClause, whereClause)
	_, err := a.DB.Exec(query, args...)
	return err
}

// Transaction executes a function within a database transaction
func (a *IncrementalAnalyzer) Transaction(fn func(tx *sql.Tx) error) error {
	tx, err := a.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// GetProgress returns the current progress from the database
func (a *IncrementalAnalyzer) GetProgress(taskID int64) (*Progress, error) {
	query := `
		SELECT processed_points, total_points, failed_points, progress_percent
		FROM analysis_tasks
		WHERE id = ?
	`

	var progress Progress
	err := a.DB.QueryRow(query, taskID).Scan(
		&progress.Processed,
		&progress.Total,
		&progress.Failed,
		&progress.Percent,
	)

	if err != nil {
		return nil, err
	}

	// Calculate ETA (simplified)
	if progress.Percent > 0 && progress.Percent < 100 {
		// Estimate based on progress rate
		// This is a simplified calculation; real implementation would track time
		remaining := progress.Total - progress.Processed
		if progress.Processed > 0 {
			progress.ETASeconds = remaining * 2 // Rough estimate: 2 seconds per record
		}
	}

	return &progress, nil
}

// MarkTaskAsCompleted marks a task as completed with optional result summary
func (a *IncrementalAnalyzer) MarkTaskAsCompleted(taskID int64, resultSummary ...string) error {
	summary := ""
	if len(resultSummary) > 0 {
		summary = resultSummary[0]
	}

	query := `
		UPDATE analysis_tasks
		SET status = 'completed',
		    progress_percent = 100,
		    result_summary = ?,
		    end_time = CURRENT_TIMESTAMP,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	_, err := a.DB.Exec(query, summary, taskID)
	return err
}

// UpdateTaskProgress updates the progress of an analysis task
func (a *IncrementalAnalyzer) UpdateTaskProgress(taskID int64, total, processed, failed int64) error {
	query := `
		UPDATE analysis_tasks
		SET processed_points = ?,
		    total_points = ?,
		    failed_points = ?,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	_, err := a.DB.Exec(query, processed, total, failed, taskID)
	return err
}

