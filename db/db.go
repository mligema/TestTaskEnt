package db

import (
	"database/sql"
	"fmt"
	"log"
)

func InitializeDatabase(connStr string) (*sql.DB, error) {
	// Open the database connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	// Initialize tables
	err = initializeTables(db)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func initializeTables(db *sql.DB) error {
	// Initialize users table if not already created exists
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS users (
		user_id BIGINT PRIMARY KEY,
		balance NUMERIC(18, 2) NOT NULL DEFAULT 0.00
	);
	`)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %v", err)
	}

	// Insert default users if they don't already exist
	_, err = db.Exec(`
	INSERT INTO users (user_id, balance)
	VALUES
		(1, 10.15),
		(2, 1.15),
		(3, 101.15)
	ON CONFLICT (user_id) DO NOTHING;
	`)
	if err != nil {
		return fmt.Errorf("failed to insert default users: %v", err)
	}

	log.Println("Database initialized with default data")
	return nil
}
