package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
	"github.com/tapiaw38/tracehub-server/internal/platform/config"
)

var db *sql.DB

// Connect establishes a connection to the PostgreSQL database
func Connect() (*sql.DB, error) {
	cfg := config.Get()
	connectionString := cfg.GetDatabaseURL()

	var err error
	db, err = sql.Open("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("✓ Database connection established")
	return db, nil
}

// GetDB returns the database connection
func GetDB() *sql.DB {
	if db == nil {
		log.Fatal("Database not initialized. Call Connect() first.")
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
