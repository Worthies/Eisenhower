package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Quadrant represents the four quadrants of the Eisenhower Matrix
type Quadrant string

const (
	UrgentImportant       Quadrant = "urgent_important"         // Do First
	NotUrgentImportant    Quadrant = "not_urgent_important"     // Schedule
	UrgentNotImportant    Quadrant = "urgent_not_important"     // Delegate
	NotUrgentNotImportant Quadrant = "not_urgent_not_important" // Eliminate
)

// Task represents a task in the Eisenhower Matrix
type Task struct {
	ID          int64      `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Quadrant    Quadrant   `json:"quadrant"`
	Priority    int        `json:"priority"` // 1-5, higher is more important
	Status      string     `json:"status"`   // pending, in_progress, completed, cancelled
	DueDate     *time.Time `json:"due_date,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	FinishedAt  *time.Time `json:"finished_at,omitempty"` // Timestamp when task was completed or cancelled
	Tags        string     `json:"tags"`                  // Comma-separated tags
	Progress    int        `json:"progress"`              // 0-100 percentage
	Summary     string     `json:"summary"`               // Progress information appended over time
}

// DB wraps the SQLite database connection
type DB struct {
	conn *sql.DB
}

// Migration represents a database schema migration
type Migration struct {
	Version int
	Name    string
	SQL     string
}

// New creates a new database connection and initializes the schema
func New(dbPath string) (*DB, error) {
	// Expand ~ to home directory
	if dbPath[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		dbPath = filepath.Join(home, dbPath[2:])
	}

	// Ensure the directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open database connection
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db := &DB{conn: conn}

	// Initialize schema
	if err := db.initSchema(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	// Run migrations
	if err := db.runMigrations(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}

// initSchema creates the tasks table if it doesn't exist
func (db *DB) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		description TEXT,
		quadrant TEXT NOT NULL CHECK(quadrant IN ('urgent_important', 'not_urgent_important', 'urgent_not_important', 'not_urgent_not_important')),
		priority INTEGER DEFAULT 3 CHECK(priority >= 1 AND priority <= 5),
		status TEXT DEFAULT 'pending' CHECK(status IN ('pending', 'in_progress', 'completed', 'cancelled')),
		due_date DATETIME,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		tags TEXT,
		progress INTEGER DEFAULT 0 CHECK(progress >= 0 AND progress <= 100),
		summary TEXT
	);

	CREATE INDEX IF NOT EXISTS idx_quadrant ON tasks(quadrant);
	CREATE INDEX IF NOT EXISTS idx_status ON tasks(status);
	CREATE INDEX IF NOT EXISTS idx_due_date ON tasks(due_date);
	`

	_, err := db.conn.Exec(schema)
	return err
}

// runMigrations runs all pending database migrations
func (db *DB) runMigrations() error {
	// Create migrations table if it doesn't exist
	migrationsTable := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	`
	if _, err := db.conn.Exec(migrationsTable); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Define all migrations
	migrations := []Migration{
		{
			Version: 1,
			Name:    "add_progress_column",
			SQL:     "ALTER TABLE tasks ADD COLUMN progress INTEGER DEFAULT 0 CHECK(progress >= 0 AND progress <= 100);",
		},
		{
			Version: 2,
			Name:    "add_summary_column",
			SQL:     "ALTER TABLE tasks ADD COLUMN summary TEXT DEFAULT '';",
		},
		{
			Version: 3,
			Name:    "add_finished_at_column",
			SQL:     "ALTER TABLE tasks ADD COLUMN finished_at DATETIME;",
		},
	}

	// Apply each migration
	for _, migration := range migrations {
		// Check if migration has already been applied
		var count int
		err := db.conn.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", migration.Version).Scan(&count)
		if err != nil {
			return fmt.Errorf("failed to check migration version %d: %w", migration.Version, err)
		}

		if count > 0 {
			// Migration already applied, skip
			continue
		}

		// Apply migration
		if _, err := db.conn.Exec(migration.SQL); err != nil {
			return fmt.Errorf("failed to apply migration %d (%s): %w", migration.Version, migration.Name, err)
		}

		// Record migration as applied
		_, err = db.conn.Exec("INSERT INTO schema_migrations (version, name) VALUES (?, ?)", migration.Version, migration.Name)
		if err != nil {
			return fmt.Errorf("failed to record migration %d: %w", migration.Version, err)
		}
	}

	return nil
}

// CreateTask creates a new task
func (db *DB) CreateTask(task *Task) error {
	query := `
		INSERT INTO tasks (title, description, quadrant, priority, status, due_date, tags, progress, summary)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := db.conn.Exec(query,
		task.Title,
		task.Description,
		task.Quadrant,
		task.Priority,
		task.Status,
		task.DueDate,
		task.Tags,
		task.Progress,
		task.Summary,
	)
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert ID: %w", err)
	}

	task.ID = id
	return nil
}

// GetTask retrieves a task by ID
func (db *DB) GetTask(id int64) (*Task, error) {
	query := `
		SELECT id, title, description, quadrant, priority, status, due_date, created_at, updated_at, finished_at, tags, progress, summary
		FROM tasks
		WHERE id = ?
	`

	task := &Task{}
	var dueDate, finishedAt sql.NullTime

	err := db.conn.QueryRow(query, id).Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&task.Quadrant,
		&task.Priority,
		&task.Status,
		&dueDate,
		&task.CreatedAt,
		&task.UpdatedAt,
		&finishedAt,
		&task.Tags,
		&task.Progress,
		&task.Summary,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("task not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	if dueDate.Valid {
		task.DueDate = &dueDate.Time
	}
	if finishedAt.Valid {
		task.FinishedAt = &finishedAt.Time
	}

	return task, nil
}

// ListTasks retrieves all tasks with optional filtering
func (db *DB) ListTasks(quadrant *Quadrant, status *string) ([]*Task, error) {
	query := `
		SELECT id, title, description, quadrant, priority, status, due_date, created_at, updated_at, finished_at, tags, progress, summary
		FROM tasks
		WHERE 1=1
	`
	args := []interface{}{}

	if quadrant != nil {
		query += " AND quadrant = ?"
		args = append(args, *quadrant)
	}

	if status != nil {
		query += " AND status = ?"
		args = append(args, *status)
	}

	query += " ORDER BY priority DESC, created_at DESC"

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		task := &Task{}
		var dueDate, finishedAt sql.NullTime

		err := rows.Scan(
			&task.ID,
			&task.Title,
			&task.Description,
			&task.Quadrant,
			&task.Priority,
			&task.Status,
			&dueDate,
			&task.CreatedAt,
			&task.UpdatedAt,
			&finishedAt,
			&task.Tags,
			&task.Progress,
			&task.Summary,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}

		if dueDate.Valid {
			task.DueDate = &dueDate.Time
		}
		if finishedAt.Valid {
			task.FinishedAt = &finishedAt.Time
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}

// UpdateTask updates an existing task
func (db *DB) UpdateTask(task *Task) error {
	query := `
		UPDATE tasks
		SET title = ?, description = ?, quadrant = ?, priority = ?, status = ?, due_date = ?, finished_at = ?, tags = ?, progress = ?, summary = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	result, err := db.conn.Exec(query,
		task.Title,
		task.Description,
		task.Quadrant,
		task.Priority,
		task.Status,
		task.DueDate,
		task.FinishedAt,
		task.Tags,
		task.Progress,
		task.Summary,
		task.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("task not found")
	}

	return nil
}

// DeleteTask deletes a task by ID
func (db *DB) DeleteTask(id int64) error {
	query := `DELETE FROM tasks WHERE id = ?`

	result, err := db.conn.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("task not found")
	}

	return nil
}

// SearchTasks searches tasks by title, description, or tags
func (db *DB) SearchTasks(searchTerm string) ([]*Task, error) {
	query := `
		SELECT id, title, description, quadrant, priority, status, due_date, created_at, updated_at, finished_at, tags, progress, summary
		FROM tasks
		WHERE title LIKE ? OR description LIKE ? OR tags LIKE ?
		ORDER BY priority DESC, created_at DESC
	`

	pattern := "%" + searchTerm + "%"
	rows, err := db.conn.Query(query, pattern, pattern, pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to search tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		task := &Task{}
		var dueDate, finishedAt sql.NullTime

		err := rows.Scan(
			&task.ID,
			&task.Title,
			&task.Description,
			&task.Quadrant,
			&task.Priority,
			&task.Status,
			&dueDate,
			&task.CreatedAt,
			&task.UpdatedAt,
			&finishedAt,
			&task.Tags,
			&task.Progress,
			&task.Summary,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}

		if dueDate.Valid {
			task.DueDate = &dueDate.Time
		}
		if finishedAt.Valid {
			task.FinishedAt = &finishedAt.Time
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}

// GetStatistics returns statistics about tasks
func (db *DB) GetStatistics() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total tasks
	var total int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM tasks").Scan(&total)
	if err != nil {
		return nil, err
	}
	stats["total"] = total

	// Tasks by quadrant
	quadrantQuery := `
		SELECT quadrant, COUNT(*) as count
		FROM tasks
		GROUP BY quadrant
	`
	rows, err := db.conn.Query(quadrantQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	byQuadrant := make(map[string]int)
	for rows.Next() {
		var quadrant string
		var count int
		if err := rows.Scan(&quadrant, &count); err != nil {
			return nil, err
		}
		byQuadrant[quadrant] = count
	}
	stats["by_quadrant"] = byQuadrant

	// Tasks by status
	statusQuery := `
		SELECT status, COUNT(*) as count
		FROM tasks
		GROUP BY status
	`
	rows, err = db.conn.Query(statusQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	byStatus := make(map[string]int)
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		byStatus[status] = count
	}
	stats["by_status"] = byStatus

	return stats, nil
}
