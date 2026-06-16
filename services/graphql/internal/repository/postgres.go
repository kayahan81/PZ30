package repository

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

type TaskRepository struct {
	db *sql.DB
}

func NewTaskRepository(host, port, user, password, dbname string) (*TaskRepository, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &TaskRepository{db: db}, nil
}

func (r *TaskRepository) Close() error {
	return r.db.Close()
}

func (r *TaskRepository) GetAll() ([]Task, error) {
	query := `SELECT id, title, description, due_date, done, created_at FROM tasks ORDER BY created_at DESC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.DueDate, &t.Done, &t.CreatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}

	return tasks, nil
}

func (r *TaskRepository) GetByID(id string) (*Task, error) {
	query := `SELECT id, title, description, due_date, done, created_at FROM tasks WHERE id = $1`

	var t Task
	err := r.db.QueryRow(query, id).Scan(&t.ID, &t.Title, &t.Description, &t.DueDate, &t.Done, &t.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &t, nil
}

func (r *TaskRepository) Create(title, description, dueDate string) (*Task, error) {
	id := fmt.Sprintf("t_%d", time.Now().UnixNano())
	now := time.Now()

	query := `
        INSERT INTO tasks (id, title, description, due_date, done, created_at)
        VALUES ($1, $2, $3, $4, $5, $6)
    `

	_, err := r.db.Exec(query, id, title, description, dueDate, false, now)
	if err != nil {
		return nil, err
	}

	return &Task{
		ID:          id,
		Title:       title,
		Description: description,
		DueDate:     dueDate,
		Done:        false,
		CreatedAt:   now,
	}, nil
}

func (r *TaskRepository) Update(id string, title, description *string, done *bool, dueDate *string) (*Task, error) {
	// Сначала получаем существующую задачу
	existing, err := r.GetByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, nil
	}

	// Обновляем только переданные поля
	if title != nil {
		existing.Title = *title
	}
	if description != nil {
		existing.Description = *description
	}
	if done != nil {
		existing.Done = *done
	}
	if dueDate != nil {
		existing.DueDate = *dueDate
	}

	query := `
        UPDATE tasks 
        SET title = $1, description = $2, done = $3, due_date = $4
        WHERE id = $5
    `

	_, err = r.db.Exec(query, existing.Title, existing.Description, existing.Done, existing.DueDate, id)
	if err != nil {
		return nil, err
	}

	return existing, nil
}

func (r *TaskRepository) Delete(id string) (bool, error) {
	query := `DELETE FROM tasks WHERE id = $1`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return rowsAffected > 0, nil
}
