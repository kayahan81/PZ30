package service

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"tasks/internal/cache"
	"tasks/internal/repository"
)

// Task - экспортируемый тип для сервисного слоя
type Task struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	DueDate     string    `json:"due_date"`
	Done        bool      `json:"done"`
	CreatedAt   time.Time `json:"created_at"`
}

type TaskService struct {
	repo  repository.TaskRepository
	cache *cache.RedisCache
}

func NewTaskService(repo repository.TaskRepository, cache *cache.RedisCache) *TaskService {
	return &TaskService{
		repo:  repo,
		cache: cache,
	}
}

// convertRepoTaskToServiceTask конвертирует repository.Task в service.Task
func convertRepoTaskToServiceTask(repoTask *repository.Task) *Task {
	if repoTask == nil {
		return nil
	}
	return &Task{
		ID:          repoTask.ID,
		Title:       repoTask.Title,
		Description: repoTask.Description,
		DueDate:     repoTask.DueDate,
		Done:        repoTask.Done,
		CreatedAt:   repoTask.CreatedAt,
	}
}

// convertServiceTaskToRepoTask конвертирует service.Task в repository.Task
func convertServiceTaskToRepoTask(serviceTask *Task) *repository.Task {
	if serviceTask == nil {
		return nil
	}
	return &repository.Task{
		ID:          serviceTask.ID,
		Title:       serviceTask.Title,
		Description: serviceTask.Description,
		DueDate:     serviceTask.DueDate,
		Done:        serviceTask.Done,
		CreatedAt:   serviceTask.CreatedAt,
	}
}

// GetTaskByID implements cache-aside pattern
func (s *TaskService) GetTaskByID(ctx context.Context, id string) (*Task, error) {
	cacheKey := s.cache.GetTaskKey(id)

	// Step 1: Try to read from Redis
	cachedData, err := s.cache.Get(ctx, cacheKey)
	if err == nil && cachedData != nil {
		// Cache hit
		var task Task
		if err := json.Unmarshal(cachedData, &task); err != nil {
			log.Printf("[WARN] Failed to unmarshal cached task %s: %v", id, err)
		} else {
			log.Printf("[INFO] Cache HIT for task %s", id)
			return &task, nil
		}
	} else if err != nil {
		log.Printf("[WARN] Redis read error for task %s, falling back to DB: %v", id, err)
	}

	log.Printf("[INFO] Cache MISS for task %s, reading from DB", id)

	// Step 2: Cache miss - read from repository (DB)
	repoTask, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Convert repository.Task to service.Task
	task := convertRepoTaskToServiceTask(repoTask)

	// Step 3: Store in cache with TTL (async to not block response)
	go func() {
		cacheCtx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		if err := s.cache.Set(cacheCtx, cacheKey, task, 0); err != nil {
			log.Printf("[WARN] Failed to cache task %s: %v", id, err)
		} else {
			log.Printf("[INFO] Cached task %s with TTL (base=%v + jitter)", id, s.cache.BaseTTL())
		}
	}()

	return task, nil
}

func (s *TaskService) GetTasks(ctx context.Context) ([]Task, error) {
	cacheKey := s.cache.GetListKey()

	cachedData, err := s.cache.Get(ctx, cacheKey)
	if err == nil && cachedData != nil {
		var tasks []Task
		if err := json.Unmarshal(cachedData, &tasks); err == nil {
			log.Printf("[INFO] Cache HIT for task list")
			return tasks, nil
		}
	}

	log.Printf("[INFO] Cache MISS for task list, reading from DB")

	repoTasks, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	// Convert repository.Tasks to service.Tasks
	tasks := make([]Task, len(repoTasks))
	for i, repoTask := range repoTasks {
		tasks[i] = Task{
			ID:          repoTask.ID,
			Title:       repoTask.Title,
			Description: repoTask.Description,
			DueDate:     repoTask.DueDate,
			Done:        repoTask.Done,
			CreatedAt:   repoTask.CreatedAt,
		}
	}

	go func() {
		cacheCtx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		if err := s.cache.Set(cacheCtx, cacheKey, tasks, 0); err != nil {
			log.Printf("[WARN] Failed to cache task list: %v", err)
		}
	}()

	return tasks, nil
}

func (s *TaskService) CreateTask(ctx context.Context, task *Task) error {
	repoTask := convertServiceTaskToRepoTask(task)
	if err := s.repo.Create(ctx, repoTask); err != nil {
		return err
	}

	// Update the original task with generated ID and CreatedAt
	task.ID = repoTask.ID
	task.CreatedAt = repoTask.CreatedAt

	// Invalidate list cache
	go func() {
		cacheCtx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		if err := s.cache.Delete(cacheCtx, s.cache.GetListKey()); err != nil {
			log.Printf("[WARN] Failed to invalidate list cache: %v", err)
		} else {
			log.Printf("[INFO] Invalidated list cache after create")
		}
	}()

	return nil
}

func (s *TaskService) UpdateTask(ctx context.Context, id string, task *Task) error {
	repoTask := convertServiceTaskToRepoTask(task)
	if err := s.repo.Update(ctx, id, repoTask); err != nil {
		return err
	}

	// Invalidate both item and list caches
	go func() {
		cacheCtx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		if err := s.cache.Delete(cacheCtx, s.cache.GetTaskKey(id)); err != nil {
			log.Printf("[WARN] Failed to invalidate task cache for %s: %v", id, err)
		}

		if err := s.cache.Delete(cacheCtx, s.cache.GetListKey()); err != nil {
			log.Printf("[WARN] Failed to invalidate list cache: %v", err)
		}

		log.Printf("[INFO] Invalidated cache for task %s and list", id)
	}()

	return nil
}

func (s *TaskService) DeleteTask(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	go func() {
		cacheCtx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		s.cache.Delete(cacheCtx, s.cache.GetTaskKey(id))
		s.cache.Delete(cacheCtx, s.cache.GetListKey())
		log.Printf("[INFO] Invalidated cache for task %s and list", id)
	}()

	return nil
}

func (s *TaskService) BaseTTL() time.Duration {
	return s.cache.BaseTTL()
}
