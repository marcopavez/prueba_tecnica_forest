package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

func Initialize(dbPath string) *sql.DB {
	// Init DB
	database, err := new(dbPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}

	if err := migrate(database); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	return database
}

func new(path string) (*sql.DB, error) {
	database, err := sql.Open("sqlite", path+"?_foreign_keys=on&_journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err := database.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}
	return database, nil
}

func migrate(database *sql.DB) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT NOT NULL UNIQUE,
			hashed_password TEXT NOT NULL,
			first_name TEXT NOT NULL,
			last_name TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS bikes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			is_available INTEGER NOT NULL DEFAULT 1,
			latitude REAL NOT NULL,
			longitude REAL NOT NULL,
			price_per_minute REAL NOT NULL DEFAULT 0.1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS rentals (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL REFERENCES users(id),
			bike_id INTEGER NOT NULL REFERENCES bikes(id),
			status TEXT NOT NULL DEFAULT 'running',
			start_time DATETIME NOT NULL,
			end_time DATETIME,
			start_latitude REAL NOT NULL,
			start_longitude REAL NOT NULL,
			end_latitude REAL,
			end_longitude REAL,
			duration_minutes INTEGER,
			cost REAL
		)`,
	}
	for _, statement := range statements {
		if _, err := database.Exec(statement); err != nil {
			return fmt.Errorf("migration error: %w", err)
		}
	}
	return nil
}
