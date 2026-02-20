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
