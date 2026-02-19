package database

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"sort"
	"strings"
)

// Migration represents a database migration
type Migration struct {
	Version int
	Name    string
	SQL     string
}

// MigrationManager manages database migrations
type MigrationManager struct {
	db             *sql.DB
	migrationsPath string
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager(db *sql.DB, migrationsPath string) *MigrationManager {
	return &MigrationManager{
		db:             db,
		migrationsPath: migrationsPath,
	}
}

// InitMigrationsTable creates the migrations tracking table
func (m *MigrationManager) InitMigrationsTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS migrations (
			version INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`
	_, err := m.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}
	return nil
}

// GetAppliedMigrations returns a list of applied migration versions
func (m *MigrationManager) GetAppliedMigrations() (map[int]bool, error) {
	rows, err := m.db.Query("SELECT version FROM migrations ORDER BY version")
	if err != nil {
		return nil, fmt.Errorf("failed to query migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[int]bool)
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("failed to scan migration version: %w", err)
		}
		applied[version] = true
	}

	return applied, nil
}

// LoadMigrations loads migration files from the migrations directory
func (m *MigrationManager) LoadMigrations() ([]Migration, error) {
	files, err := ioutil.ReadDir(m.migrationsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	var migrations []Migration
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}

		// Parse version from filename (e.g., "001_add_admin_columns.sql")
		var version int
		var name string
		_, err := fmt.Sscanf(file.Name(), "%d_%s", &version, &name)
		if err != nil {
			log.Printf("Warning: skipping migration file with invalid name: %s", file.Name())
			continue
		}

		// Read migration SQL
		content, err := ioutil.ReadFile(filepath.Join(m.migrationsPath, file.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed to read migration file %s: %w", file.Name(), err)
		}

		migrations = append(migrations, Migration{
			Version: version,
			Name:    strings.TrimSuffix(file.Name(), ".sql"),
			SQL:     string(content),
		})
	}

	// Sort migrations by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// ApplyMigration applies a single migration
func (m *MigrationManager) ApplyMigration(migration Migration) error {
	tx, err := m.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	// Execute migration SQL
	_, err = tx.Exec(migration.SQL)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to execute migration %d: %w", migration.Version, err)
	}

	// Record migration
	_, err = tx.Exec("INSERT INTO migrations (version, name) VALUES (?, ?)", migration.Version, migration.Name)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to record migration %d: %w", migration.Version, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration %d: %w", migration.Version, err)
	}

	log.Printf("Applied migration %d: %s", migration.Version, migration.Name)
	return nil
}

// RunMigrations runs all pending migrations
func (m *MigrationManager) RunMigrations() error {
	// Initialize migrations table
	if err := m.InitMigrationsTable(); err != nil {
		return err
	}

	// Get applied migrations
	applied, err := m.GetAppliedMigrations()
	if err != nil {
		return err
	}

	// Load migrations
	migrations, err := m.LoadMigrations()
	if err != nil {
		return err
	}

	// Apply pending migrations
	for _, migration := range migrations {
		if applied[migration.Version] {
			log.Printf("Skipping already applied migration %d: %s", migration.Version, migration.Name)
			continue
		}

		if err := m.ApplyMigration(migration); err != nil {
			return err
		}
	}

	log.Println("All migrations applied successfully")
	return nil
}
