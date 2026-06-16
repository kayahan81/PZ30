package graph

import (
	"context"
	"fmt"

	"tech-ip-sem2/services/graphql/graph/model"
	"tech-ip-sem2/services/graphql/internal/repository"
)

// ============================================
// QUERY резолверы
// ============================================

// Tasks возвращает список всех задач
func (r *queryResolver) Tasks(ctx context.Context) ([]*repository.Task, error) {
	tasks, err := r.Repo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks: %w", err)
	}

	// Конвертируем []Task в []*Task
	result := make([]*repository.Task, len(tasks))
	for i := range tasks {
		result[i] = &tasks[i]
	}
	return result, nil
}

// Task возвращает задачу по ID
func (r *queryResolver) Task(ctx context.Context, id string) (*repository.Task, error) {
	task, err := r.Repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	return task, nil
}

// ============================================
// MUTATION резолверы
// ============================================

// CreateTask создаёт новую задачу
func (r *mutationResolver) CreateTask(ctx context.Context, input model.CreateTaskInput) (*repository.Task, error) {
	// Если dueDate не передан — оставляем пустым
	dueDate := ""
	if input.DueDate != nil {
		dueDate = *input.DueDate
	}

	task, err := r.Repo.Create(input.Title, input.Description, dueDate)
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}
	return task, nil
}

// UpdateTask обновляет задачу
func (r *mutationResolver) UpdateTask(ctx context.Context, id string, input model.UpdateTaskInput) (*repository.Task, error) {
	task, err := r.Repo.Update(
		id,
		input.Title,
		input.Description,
		input.Done,
		input.DueDate,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}
	if task == nil {
		return nil, fmt.Errorf("task with id %s not found", id)
	}
	return task, nil
}

// DeleteTask удаляет задачу
func (r *mutationResolver) DeleteTask(ctx context.Context, id string) (bool, error) {
	deleted, err := r.Repo.Delete(id)
	if err != nil {
		return false, fmt.Errorf("failed to delete task: %w", err)
	}
	return deleted, nil
}

// ============================================
// Генерация кода (необходимо для компиляции)
// ============================================

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

// Mutation returns MutationResolver implementation.
func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }

type queryResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }
