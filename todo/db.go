package main

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

var db *sql.DB

func getDBPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".todo.db")
}

func InitDB() error {
	var err error
	db, err = sql.Open("sqlite", getDBPath())
	if err != nil {
		return err
	}

	// Todo schema
	schema := `
	CREATE TABLE IF NOT EXISTS todos (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		task TEXT NOT NULL,
		done BOOLEAN DEFAULT FALSE,
		priority TEXT DEFAULT 'medium',
		category TEXT DEFAULT '',
		due_date TEXT DEFAULT '',
		created_at TEXT,
		updated_at TEXT
	);
	`
	_, err = db.Exec(schema)
	if err != nil {
		return err
	}

	// Vault schema
	return InitVaultDB()
}

func CloseDB() {
	if db != nil {
		db.Close()
	}
}

func CreateTodo(task string, priority Priority, category string, dueDate string) (*Todo, error) {
	now := time.Now()
	result, err := db.Exec(
		`INSERT INTO todos (task, done, priority, category, due_date, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		task, false, priority, category, dueDate, now.Format(time.RFC3339), now.Format(time.RFC3339),
	)
	if err != nil {
		return nil, err
	}

	id, _ := result.LastInsertId()
	return &Todo{
		ID:        id,
		Task:      task,
		Done:      false,
		Priority:  priority,
		Category:  category,
		DueDate:   dueDate,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func GetTodos(filter TodoFilter) ([]Todo, error) {
	query := `SELECT id, task, done, priority, category, due_date, created_at, updated_at FROM todos WHERE 1=1`
	args := []interface{}{}

	if filter.Status == "done" {
		query += " AND done = ?"
		args = append(args, true)
	} else if filter.Status == "pending" {
		query += " AND done = ?"
		args = append(args, false)
	}

	if filter.Priority != "" && filter.Priority != "all" {
		query += " AND priority = ?"
		args = append(args, filter.Priority)
	}

	if filter.Category != "" {
		query += " AND category = ?"
		args = append(args, filter.Category)
	}

	if filter.Search != "" {
		query += " AND task LIKE ?"
		args = append(args, "%"+filter.Search+"%")
	}

	query += " ORDER BY done ASC, CASE priority WHEN 'high' THEN 1 WHEN 'medium' THEN 2 WHEN 'low' THEN 3 END, created_at DESC"

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var t Todo
		var createdAt, updatedAt string
		var priority string
		err := rows.Scan(&t.ID, &t.Task, &t.Done, &priority, &t.Category, &t.DueDate, &createdAt, &updatedAt)
		if err != nil {
			continue
		}
		t.Priority = Priority(priority)
		t.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		t.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
		todos = append(todos, t)
	}
	return todos, nil
}

func GetTodo(id int64) (*Todo, error) {
	var t Todo
	var createdAt, updatedAt string
	var priority string
	err := db.QueryRow(
		`SELECT id, task, done, priority, category, due_date, created_at, updated_at FROM todos WHERE id = ?`,
		id,
	).Scan(&t.ID, &t.Task, &t.Done, &priority, &t.Category, &t.DueDate, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}
	t.Priority = Priority(priority)
	t.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	t.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	return &t, nil
}

func UpdateTodo(id int64, task string, done bool, priority Priority, category string, dueDate string) error {
	now := time.Now()
	_, err := db.Exec(
		`UPDATE todos SET task=?, done=?, priority=?, category=?, due_date=?, updated_at=? WHERE id=?`,
		task, done, priority, category, dueDate, now.Format(time.RFC3339), id,
	)
	return err
}

func MarkTodoDone(id int64, done bool) error {
	now := time.Now()
	_, err := db.Exec(`UPDATE todos SET done=?, updated_at=? WHERE id=?`, done, now.Format(time.RFC3339), id)
	return err
}

func DeleteTodo(id int64) error {
	_, err := db.Exec(`DELETE FROM todos WHERE id=?`, id)
	return err
}

func ClearTodos() error {
	_, err := db.Exec(`DELETE FROM todos`)
	return err
}

func GetCategories() ([]string, error) {
	rows, err := db.Query(`SELECT DISTINCT category FROM todos WHERE category != '' ORDER BY category`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []string
	for rows.Next() {
		var cat string
		rows.Scan(&cat)
		if strings.TrimSpace(cat) != "" {
			categories = append(categories, cat)
		}
	}
	return categories, nil
}
