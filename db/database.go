package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ashwinath/reminder/models"

	_ "modernc.org/sqlite"
)

type Database struct {
	conn *sql.DB
}

func New() (*Database, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	dbDir := filepath.Join(homeDir, ".reminder")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	dbPath := filepath.Join(dbDir, "db.sqlite")
	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	database := &Database{conn: conn}
	if err := database.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return database, nil
}

func (d *Database) initSchema() error {
	query := `
	CREATE TABLE IF NOT EXISTS reminders (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		status TEXT NOT NULL DEFAULT 'active',
		description TEXT NOT NULL,
		url TEXT NOT NULL,
		remarks TEXT NOT NULL DEFAULT '',
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);`
	_, err := d.conn.Exec(query)
	if err != nil {
		return err
	}

	_, err = d.conn.Exec("ALTER TABLE reminders ADD COLUMN remarks TEXT NOT NULL DEFAULT ''")
	return nil
}

func (d *Database) Add(description, url string) (*models.Reminder, error) {
	now := time.Now()
	result, err := d.conn.Exec(
		"INSERT INTO reminders (status, description, url, remarks, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
		"active", description, url, "", now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add reminder: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return &models.Reminder{
		ID:          id,
		Status:      "active",
		Description: description,
		URL:         url,
		Remarks:     "",
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

func (d *Database) Complete(ids []int64) ([]models.Reminder, error) {
	var reminders []models.Reminder
	now := time.Now()

	for _, id := range ids {
		result, err := d.conn.Exec(
			"UPDATE reminders SET status = ?, updated_at = ? WHERE id = ?",
			"completed", now, id,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to complete reminder %d: %w", id, err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return nil, fmt.Errorf("failed to check rows affected: %w", err)
		}

		if rowsAffected == 0 {
			return nil, fmt.Errorf("reminder with ID %d not found", id)
		}

		reminder, err := d.GetByID(id)
		if err != nil {
			return nil, fmt.Errorf("failed to get reminder %d: %w", id, err)
		}
		reminders = append(reminders, *reminder)
	}

	return reminders, nil
}

func (d *Database) Activate(ids []int64) ([]models.Reminder, error) {
	var reminders []models.Reminder
	now := time.Now()

	for _, id := range ids {
		result, err := d.conn.Exec(
			"UPDATE reminders SET status = ?, updated_at = ? WHERE id = ?",
			"active", now, id,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to activate reminder %d: %w", id, err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return nil, fmt.Errorf("failed to check rows affected: %w", err)
		}

		if rowsAffected == 0 {
			return nil, fmt.Errorf("reminder with ID %d not found", id)
		}

		reminder, err := d.GetByID(id)
		if err != nil {
			return nil, fmt.Errorf("failed to get reminder %d: %w", id, err)
		}
		reminders = append(reminders, *reminder)
	}

	return reminders, nil
}

func (d *Database) Delete(ids []int64) ([]models.Reminder, error) {
	var reminders []models.Reminder

	for _, id := range ids {
		reminder, err := d.GetByID(id)
		if err != nil {
			return nil, fmt.Errorf("failed to get reminder %d: %w", id, err)
		}
		reminders = append(reminders, *reminder)

		_, err = d.conn.Exec("DELETE FROM reminders WHERE id = ?", id)
		if err != nil {
			return nil, fmt.Errorf("failed to delete reminder %d: %w", id, err)
		}
	}

	return reminders, nil
}

func (d *Database) GetByID(id int64) (*models.Reminder, error) {
	reminder := &models.Reminder{}
	err := d.conn.QueryRow(
		"SELECT id, status, description, url, remarks, created_at, updated_at FROM reminders WHERE id = ?",
		id,
	).Scan(&reminder.ID, &reminder.Status, &reminder.Description, &reminder.URL, &reminder.Remarks, &reminder.CreatedAt, &reminder.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("reminder with ID %d not found", id)
	}

	return reminder, nil
}

func (d *Database) GetAll(status string) ([]models.Reminder, error) {
	var query string
	var args []any

	if status == "all" {
		query = "SELECT id, status, description, url, remarks, created_at, updated_at FROM reminders ORDER BY id"
	} else {
		query = "SELECT id, status, description, url, remarks, created_at, updated_at FROM reminders WHERE status = ? ORDER BY id"
		args = append(args, "active")
	}

	rows, err := d.conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query reminders: %w", err)
	}
	defer rows.Close()

	var reminders []models.Reminder
	for rows.Next() {
		var r models.Reminder
		if err := rows.Scan(&r.ID, &r.Status, &r.Description, &r.URL, &r.Remarks, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan reminder: %w", err)
		}
		reminders = append(reminders, r)
	}

	return reminders, nil
}

func (d *Database) Cancel(id int64, reason string) (*models.Reminder, error) {
	now := time.Now()

	result, err := d.conn.Exec(
		"UPDATE reminders SET status = ?, remarks = ?, updated_at = ? WHERE id = ?",
		"cancelled", reason, now, id,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to cancel reminder %d: %w", id, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return nil, fmt.Errorf("reminder with ID %d not found", id)
	}

	return d.GetByID(id)
}

func (d *Database) Close() error {
	return d.conn.Close()
}
