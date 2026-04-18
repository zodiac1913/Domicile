package handlers

import (
	"Domicile/internal/database"
	"Domicile/templates/partials"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/labstack/echo/v4"
)

const dataDir = "data"
const dbFileName = "domicile.db"

// BackupDownload creates a gzipped backup of the database and sends it as a download.
func BackupDownload(c echo.Context) error {
	dbPath := filepath.Join(dataDir, dbFileName)

	// Checkpoint WAL to make sure all data is in the main DB file.
	database.DB.Exec("PRAGMA wal_checkpoint(TRUNCATE)")

	// Open the DB file.
	f, err := os.Open(dbPath)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to open database: "+err.Error())
	}
	defer f.Close()

	// Set headers for download.
	timestamp := time.Now().Format("2006-01-02_150405")
	filename := fmt.Sprintf("domicile_backup_%s.db.gz", timestamp)
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Response().Header().Set("Content-Type", "application/gzip")
	c.Response().WriteHeader(http.StatusOK)

	// Gzip compress directly to the response.
	gz, err := gzip.NewWriterLevel(c.Response(), gzip.BestCompression)
	if err != nil {
		return err
	}
	defer gz.Close()

	gz.Name = dbFileName
	gz.ModTime = time.Now()

	_, err = io.Copy(gz, f)
	return err
}

// BackupRestore replaces the database with an uploaded gzipped backup.
func BackupRestore(c echo.Context) error {
	file, err := c.FormFile("backup")
	if err != nil {
		return c.String(http.StatusBadRequest, "No backup file provided")
	}

	src, err := file.Open()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to read upload: "+err.Error())
	}
	defer src.Close()

	// Decompress gzip.
	gz, err := gzip.NewReader(src)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid backup file — not a valid gzip archive")
	}
	defer gz.Close()

	// Write to a temp file first.
	dbPath := filepath.Join(dataDir, dbFileName)
	tmpPath := dbPath + ".restore.tmp"

	tmp, err := os.Create(tmpPath)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to create temp file: "+err.Error())
	}

	_, err = io.Copy(tmp, gz)
	tmp.Close()
	if err != nil {
		os.Remove(tmpPath)
		return c.String(http.StatusInternalServerError, "Failed to decompress backup: "+err.Error())
	}

	// Validate it's a real SQLite DB by trying to open it.
	testDB, err := os.Open(tmpPath)
	if err != nil {
		os.Remove(tmpPath)
		return c.String(http.StatusBadRequest, "Invalid backup file")
	}
	header := make([]byte, 16)
	testDB.Read(header)
	testDB.Close()
	if string(header[:13]) != "SQLite format" {
		os.Remove(tmpPath)
		return c.String(http.StatusBadRequest, "Invalid backup file — not a SQLite database")
	}

	// Close current DB connection.
	database.Close()

	// Remove WAL and SHM files.
	os.Remove(dbPath + "-wal")
	os.Remove(dbPath + "-shm")

	// Replace the DB file.
	err = os.Rename(tmpPath, dbPath)
	if err != nil {
		// Try to reopen the old DB.
		database.Init(dataDir)
		return c.String(http.StatusInternalServerError, "Failed to replace database: "+err.Error())
	}

	// Reopen the database.
	err = database.Init(dataDir)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Database restored but failed to reopen: "+err.Error())
	}

	// Redirect back to settings.
	c.Response().Header().Set("HX-Redirect", "/settings")
	return c.NoContent(http.StatusOK)
}

// BackupSettingsPage renders the backup/restore settings panel.
func BackupSettingsPage(c echo.Context) error {
	return render(c, http.StatusOK, partials.BackupRestore())
}
