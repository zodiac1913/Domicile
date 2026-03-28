package database

import (
	_ "embed"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaSQL string

var DB *sqlx.DB

// Init opens (or creates) the SQLite database and runs the schema.
func Init(dataDir string) error {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("create data dir: %w", err)
	}

	dbPath := filepath.Join(dataDir, "domicile.db")
	var err error
	DB, err = sqlx.Open("sqlite", dbPath+"?_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)")
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}

	// SQLite performs best with a single writer
	DB.SetMaxOpenConns(1)

	if _, err := DB.Exec(schemaSQL); err != nil {
		return fmt.Errorf("apply schema: %w", err)
	}

	log.Printf("Database ready at %s", dbPath)
	return nil
}

// Close cleanly shuts down the database connection.
func Close() {
	if DB != nil {
		DB.Close()
	}
}
