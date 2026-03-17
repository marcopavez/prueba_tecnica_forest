package db

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

func New(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path+"?_foreign_keys=on&_journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}
	return db, nil
}

func Migrate(db *sql.DB) error {
	stmts := []string{
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
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			return fmt.Errorf("migration error: %w", err)
		}
	}
	return nil
}
