package repository

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Task struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	DueDate     string    `json:"due_date"`
	Done        bool      `json:"done"`
	CreatedAt   time.Time `json:"created_at"`
}

type TaskRepository interface {
	GetByID(ctx context.Context, id string) (*Task, error)
	GetAll(ctx context.Context) ([]Task, error)
	Create(ctx context.Context, task *Task) error
	Update(ctx context.Context, id string, task *Task) error
	Delete(ctx context.Context, id string) error
}

type InMemoryRepository struct {
	mu      sync.RWMutex
	tasks   map[string]Task
	counter int
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		tasks:   make(map[string]Task),
		counter: 0,
	}
}

func (r *InMemoryRepository) nextID() string {
	r.counter++
	return fmt.Sprintf("t_%03d", r.counter)
}

func (r *InMemoryRepository) GetByID(ctx context.Context, id string) (*Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	task, exists := r.tasks[id]
	if !exists {
		return nil, fmt.Errorf("task not found: %s", id)
	}

	return &task, nil
}

func (r *InMemoryRepository) GetAll(ctx context.Context) ([]Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tasks := make([]Task, 0, len(r.tasks))
	for _, task := range r.tasks {
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (r *InMemoryRepository) Create(ctx context.Context, task *Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	task.ID = r.nextID()
	task.CreatedAt = time.Now()
	r.tasks[task.ID] = *task
	return nil
}

func (r *InMemoryRepository) Update(ctx context.Context, id string, task *Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, exists := r.tasks[id]
	if !exists {
		return fmt.Errorf("task not found: %s", id)
	}

	task.ID = id
	r.tasks[id] = *task
	return nil
}

func (r *InMemoryRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, exists := r.tasks[id]
	if !exists {
		return fmt.Errorf("task not found: %s", id)
	}

	delete(r.tasks, id)
	return nil
}
