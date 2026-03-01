package database

import (
	"database/sql"
	"fmt"
	"log"
	"sync"

	_ "modernc.org/sqlite"
)

var (
	db   *sql.DB
	once sync.Once
)

// Config holds database configuration
type Config struct {
	Path string
}

// Init initializes the database connection
func Init(cfg Config) error {
	var err error
	once.Do(func() {
		db, err = sql.Open("sqlite", cfg.Path)
		if err != nil {
			return
		}

		// Set connection pool settings
		db.SetMaxOpenConns(10)
		db.SetMaxIdleConns(5)

		// Enable WAL mode for better concurrency
		_, err = db.Exec("PRAGMA journal_mode=WAL")
		if err != nil {
			return
		}

		// Set busy timeout to 30 seconds to handle concurrent access
		_, err = db.Exec("PRAGMA busy_timeout=30000")
		if err != nil {
			return
		}

		// Enable foreign keys
		_, err = db.Exec("PRAGMA foreign_keys=ON")
		if err != nil {
			return
		}

		// Test connection
		err = db.Ping()
		if err != nil {
			return
		}

		log.Printf("Database initialized successfully: %s", cfg.Path)
	})

	return err
}

// GetDB returns the database instance
func GetDB() *sql.DB {
	if db == nil {
		log.Fatal("Database not initialized. Call Init() first.")
	}
	return db
}

// Close closes the database connection
func Close() error {
	if db != nil {
		return db.Close()
	}
	return nil
}

// OpenDB opens a new database connection (for additional databases like flights, keyboard, etc.)
func OpenDB(path string) (*sql.DB, error) {
	newDB, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	newDB.SetMaxOpenConns(10)
	newDB.SetMaxIdleConns(5)

	// Enable WAL mode for better concurrency
	if _, err := newDB.Exec("PRAGMA journal_mode=WAL"); err != nil {
		newDB.Close()
		return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
	}

	// Set busy timeout to 30 seconds
	if _, err := newDB.Exec("PRAGMA busy_timeout=30000"); err != nil {
		newDB.Close()
		return nil, fmt.Errorf("failed to set busy timeout: %w", err)
	}

	// Enable foreign keys
	if _, err := newDB.Exec("PRAGMA foreign_keys=ON"); err != nil {
		newDB.Close()
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Test connection
	if err := newDB.Ping(); err != nil {
		newDB.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("Database opened successfully: %s", path)
	return newDB, nil
}

// Transaction executes a function within a database transaction
func Transaction(fn func(*sql.Tx) error) error {
	tx, err := db.Begin()
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
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("transaction error: %v, rollback error: %w", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
